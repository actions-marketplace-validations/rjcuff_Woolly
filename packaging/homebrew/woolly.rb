class Woolly < Formula
  desc "Local project memory for coding agents that flags stale facts"
  homepage "https://usewoolly.com"
  version "0.20.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.0/woolly_darwin_arm64.tar.gz"
      sha256 "30f2f1e109c8026a7ad02d413988af8f3057bd02a3c9622e13e8f2bcdbf2e1ec"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.0/woolly_darwin_amd64.tar.gz"
      sha256 "3bec6d34d84feecb4114b54cae33ae0496749b04f4697386fdbaf058863fdc56"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.0/woolly_linux_arm64.tar.gz"
      sha256 "f014ff810bb80a5abe0af0bb6db82d6f93d5d0fa2ec16a988c67be3208e24ca5"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.0/woolly_linux_amd64.tar.gz"
      sha256 "aa89aadad5f394f0c9ef44feb34d1d3c209c7d1ab4ca230860ad20d74db5de8a"
    end
  end

  def install
    bin.install "woolly"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/woolly version")
    ENV["WOOLLY_HOME"] = testpath/"woolly"
    ENV["WOOLLY_NO_EMBED"] = "1"
    system bin/"woolly", "remember", "Homebrew test"
    assert_match "Homebrew test", shell_output("#{bin}/woolly list")
  end
end
