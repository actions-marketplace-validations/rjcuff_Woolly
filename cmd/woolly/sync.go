package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjcuff/woolly/internal/activity"
	"github.com/rjcuff/woolly/internal/peer"
)

const syncUsage = `Keep your memories in step between your own devices.

  woolly sync init                 set up sync on this device and print a join command
  woolly sync join <addr> <code>   link this device to another one
  woolly sync                      sync with your other devices now
  woolly sync folder <path>        sync through a folder you already sync, like Syncthing or a git repo
  woolly sync folder <path> --code <code>   use that folder from another device
  woolly sync folder <path> --remove        stop syncing through that folder
  woolly sync status [--show-code]  list linked devices and folders
  woolly sync serve                listen for your other devices (woolly serve does this too)
  woolly sync leave                stop syncing on this device
`

// syncInterestURL is where people can say they'd use sync that needs no
// setup. It's only a link: Woolly sends nothing.
const syncInterestURL = "https://github.com/rjcuff/woolly/issues/new?title=I%27d+use+sync+away+from+home"

func cmdSync(args []string) error {
	if len(args) == 0 {
		return syncNow()
	}
	switch args[0] {
	case "init":
		return syncInit()
	case "join":
		if len(args) != 3 {
			return errors.New("usage: woolly sync join <address> <code>")
		}
		return syncJoin(args[1], args[2])
	case "folder":
		return syncFolderCmd(args[1:])
	case "status":
		return syncStatus(len(args) > 1 && args[1] == "--show-code")
	case "serve":
		return syncServe()
	case "leave":
		return syncLeave()
	case "help", "-h", "--help":
		fmt.Print(syncUsage)
		return nil
	}
	return fmt.Errorf("unknown sync command %q\n\n%s", args[0], syncUsage)
}

func syncAddr() string {
	return envOr("WOOLLY_SYNC_ADDR", "0.0.0.0:"+peer.DefaultPort)
}

func syncInit() error {
	a, err := open()
	if err != nil {
		return err
	}
	cfg, err := peer.Load(a.home)
	if err != nil {
		return err
	}
	if cfg == nil {
		if cfg, err = peer.NewConfig(); err != nil {
			return err
		}
	}
	// Other devices will join over the network, so listen for them.
	cfg.FolderOnly = false
	if err := cfg.Save(a.home); err != nil {
		return err
	}
	ip := "<this-device's-address>"
	if addrs := peer.LocalAddrs(); len(addrs) > 0 {
		ip = addrs[0]
	}
	fmt.Printf(`Sync is set up on this device.

On your other device, run:

  woolly sync join %s %s

Keep the code private. Anyone with it can read and change your memories.
This device needs to be running Woolly (woolly serve or woolly service install)
for the other one to reach it.
`, ip, cfg.JoinCode())
	return nil
}

func syncJoin(addr, code string) error {
	a, err := open()
	if err != nil {
		return err
	}
	cfg, err := peer.Load(a.home)
	if err != nil {
		return err
	}
	if cfg == nil {
		if cfg, err = peer.NewConfig(); err != nil {
			return err
		}
	}
	if !peer.ValidCode(code) {
		return errors.New("that code doesn't look right. Copy it again from `woolly sync init`")
	}
	if cfg.Key != code {
		if len(cfg.Peers)+len(cfg.Folders) > 0 {
			return errors.New("this device already syncs with a different group. Run `woolly sync leave` first")
		}
		cfg.Key = code
	}
	cfg.FolderOnly = false
	addr = peer.NormalizeAddr(addr)
	cfg.AddPeer(addr)
	if err := cfg.Save(a.home); err != nil {
		return err
	}
	fmt.Println("Linked with", addr)
	return syncNow()
}

func syncNow() error {
	a, err := open()
	if err != nil {
		return err
	}
	cfg, err := peer.Load(a.home)
	if err != nil {
		return err
	}
	if cfg == nil || len(cfg.Peers)+len(cfg.Folders) == 0 {
		fmt.Print("Sync isn't set up yet.\n\n" + syncUsage)
		return nil
	}
	failed := syncAll(a, cfg, true)
	if failed > 0 {
		return fmt.Errorf("%d of %d devices and folders didn't sync", failed, len(cfg.Peers)+len(cfg.Folders))
	}
	return nil
}

// syncFolderCmd adds or removes a sync folder.
func syncFolderCmd(args []string) error {
	fs := newFlags("sync folder")
	code := fs.String("code", "", "the join code from your first device, when this folder already has your memories")
	remove := fs.Bool("remove", false, "stop syncing through this folder; its files are left alone")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return errors.New("usage: woolly sync folder <path> [--code <code>] [--remove]")
	}
	path, err := filepath.Abs(pos[0])
	if err != nil {
		return err
	}
	a, err := open()
	if err != nil {
		return err
	}
	cfg, err := peer.Load(a.home)
	if err != nil {
		return err
	}
	if *remove {
		if cfg == nil || !cfg.RemoveFolder(path) {
			return fmt.Errorf("%s isn't a sync folder", path)
		}
		if err := cfg.Save(a.home); err != nil {
			return err
		}
		fmt.Println("Stopped syncing through", path+". Its files are still there.")
		return nil
	}
	if st, err := os.Stat(path); err != nil || !st.IsDir() {
		return fmt.Errorf("%s isn't a folder. Create it, or point at the folder your sync tool shares", path)
	}
	if cfg == nil {
		if cfg, err = peer.NewConfig(); err != nil {
			return err
		}
		cfg.FolderOnly = true
	}
	if *code != "" {
		if !peer.ValidCode(*code) {
			return errors.New("that code doesn't look right. Copy it again from `woolly sync status --show-code` on your first device")
		}
		if cfg.Key != *code && len(cfg.Peers)+len(cfg.Folders) > 0 {
			return errors.New("this device already syncs with a different code. Run `woolly sync leave` first")
		}
		cfg.Key = *code
	}
	cfg.AddFolder(path)
	if err := cfg.Save(a.home); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	res, err := peer.SyncFolder(ctx, a.store, cfg, path)
	cancel()
	if err != nil {
		cfg.RemoveFolder(path)
		_ = cfg.Save(a.home)
		return err
	}
	for i := range cfg.Folders {
		if filepath.Clean(cfg.Folders[i].Path) == filepath.Clean(path) || strings.EqualFold(cfg.Folders[i].Path, path) {
			cfg.Folders[i].LastSync, cfg.Folders[i].LastErr = time.Now().UTC(), ""
		}
	}
	if err := cfg.Save(a.home); err != nil {
		return err
	}
	fmt.Printf("Syncing through %s. Got %d, wrote %d.\n\n", path, res.Received, res.Sent)
	fmt.Printf(`Woolly writes encrypted files to %s and checks it every 5 minutes
while Woolly runs (woolly serve or woolly service install). Moving the folder
between computers is up to the tool that shares it, like Syncthing, Dropbox
or a git repo you push and pull.

On your other devices, run:

  woolly sync folder <path to the same folder> --code %s

Keep the code private. Anyone with it and the folder can read your memories.
`, filepath.Join(path, peer.FolderDir), cfg.JoinCode())
	return nil
}

// syncAll syncs with every peer, records the result and returns how many failed.
func syncAll(a *app, cfg *peer.Config, verbose bool) int {
	client := &peer.Client{}
	failed := 0
	for i := range cfg.Peers {
		p := &cfg.Peers[i]
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		res, err := client.Sync(ctx, a.store, cfg, p.Addr)
		cancel()
		if err != nil {
			failed++
			p.LastErr = err.Error()
			if verbose {
				fmt.Printf("  ! %s: %v\n", p.Addr, err)
			}
			continue
		}
		p.LastErr, p.LastSync = "", time.Now().UTC()
		if res.Name != "" {
			p.Name = res.Name
		}
		if verbose {
			fmt.Printf("  ✓ %s: got %d, sent %d", label(*p), res.Received, res.Sent)
			if res.Deleted > 0 {
				fmt.Printf(", deleted %d", res.Deleted)
			}
			fmt.Println()
		}
		if res.Received+res.Sent+res.Deleted > 0 {
			a.log.Add(activity.Entry{App: label(*p), Action: "sync", Count: res.Received + res.Sent + res.Deleted})
		}
	}
	for i := range cfg.Folders {
		f := &cfg.Folders[i]
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		res, err := peer.SyncFolder(ctx, a.store, cfg, f.Path)
		cancel()
		if err != nil {
			failed++
			f.LastErr = err.Error()
			if verbose {
				fmt.Printf("  ! %s: %v\n", f.Path, err)
			}
			continue
		}
		f.LastErr, f.LastSync = "", time.Now().UTC()
		if verbose {
			fmt.Printf("  ✓ %s: got %d, wrote %d", f.Path, res.Received, res.Sent)
			if res.Deleted > 0 {
				fmt.Printf(", deleted %d", res.Deleted)
			}
			if res.Skipped > 0 {
				fmt.Printf(", skipped %d that didn't open with your code", res.Skipped)
			}
			fmt.Println()
		}
		if res.Received+res.Sent+res.Deleted > 0 {
			a.log.Add(activity.Entry{App: "sync folder", Action: "sync", Count: res.Received + res.Sent + res.Deleted})
		}
	}
	_ = cfg.Save(a.home)
	return failed
}

func label(p peer.Peer) string {
	if p.Name != "" {
		return p.Name + " (" + p.Addr + ")"
	}
	return p.Addr
}

// syncStatus lists devices and folders. The join code is the key to every
// sync folder, so it's printed only when asked for.
func syncStatus(showCode bool) error {
	a, err := open()
	if err != nil {
		return err
	}
	cfg, err := peer.Load(a.home)
	if err != nil {
		return err
	}
	if cfg == nil {
		fmt.Println("Sync isn't set up. Run: woolly sync init")
		return nil
	}
	fmt.Println("This device:", cfg.Name)
	if listens(cfg) {
		fmt.Println("Listens on:", syncAddr())
	}
	if len(cfg.Peers)+len(cfg.Folders) == 0 {
		fmt.Println("No linked devices yet. On another device, run the join command from `woolly sync init`.")
	}
	for _, p := range cfg.Peers {
		printSyncRow(label(p), p.LastSync, p.LastErr)
	}
	if len(cfg.Folders) > 0 {
		fmt.Println("Sync folders:")
		for _, f := range cfg.Folders {
			printSyncRow(f.Path, f.LastSync, f.LastErr)
		}
	}
	switch {
	case showCode:
		fmt.Println("\nJoin code:", cfg.JoinCode())
		fmt.Println("Keep it private. Anyone with it can read your memories.")
	case len(cfg.Folders) > 0:
		fmt.Println("\nTo set up another device, run: woolly sync status --show-code")
	}
	fmt.Println("\nWant sync away from home with no setup? Tell us:", syncInterestURL)
	return nil
}

func printSyncRow(name string, last time.Time, lastErr string) {
	when := "never"
	if !last.IsZero() {
		when = last.Local().Format("Jan 02 15:04")
	}
	fmt.Printf("  %s  last sync: %s\n", name, when)
	if lastErr != "" {
		fmt.Printf("    last error: %s\n", lastErr)
	}
}

// listens reports whether this device listens for others on the network.
// A device that only syncs through folders doesn't need to.
func listens(cfg *peer.Config) bool {
	return !cfg.FolderOnly
}

func syncLeave() error {
	a, err := open()
	if err != nil {
		return err
	}
	if err := os.Remove(peer.ConfigPath(a.home)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	fmt.Println("Sync is off on this device. Your memories are still here.")
	return nil
}

func syncServe() error {
	a, err := open()
	if err != nil {
		return err
	}
	stop, err := startSync(a, true)
	if err != nil {
		return err
	}
	if stop == nil {
		return errors.New("sync isn't set up. Run: woolly sync init")
	}
	select {}
}

// ensureSync starts the sync listener once, if sync is set up.
func (a *app) ensureSync() error {
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.syncStop != nil {
		return nil
	}
	stop, err := startSync(a, false)
	a.syncStop = stop
	return err
}

// startSync listens for other devices and syncs every 5 minutes. It returns
// nil when sync isn't set up.
func startSync(a *app, verbose bool) (func(), error) {
	cfg, err := peer.Load(a.home)
	if err != nil || cfg == nil {
		return nil, err
	}
	var srv *http.Server
	if listens(cfg) {
		h, err := peer.Handler(a.store, cfg)
		if err != nil {
			return nil, err
		}
		ln, err := net.Listen("tcp", syncAddr())
		if err != nil {
			return nil, fmt.Errorf("can't listen for other devices on %s: %w", syncAddr(), err)
		}
		srv = &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
		go srv.Serve(ln)
		if verbose {
			fmt.Fprintf(os.Stderr, "Listening for your other devices on %s\n", syncAddr())
		}
	}
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			if c, err := peer.Load(a.home); err == nil && c != nil && len(c.Peers)+len(c.Folders) > 0 {
				syncAll(a, c, false)
			}
			select {
			case <-t.C:
			case <-done:
				return
			}
		}
	}()
	return func() {
		close(done)
		if srv != nil {
			srv.Close()
		}
	}, nil
}
