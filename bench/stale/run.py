"""Stale-fact benchmark for Woolly.

    python3 bench/stale/run.py

Builds a small ESP32 drone firmware repo, saves a fact about its code, makes
one change and commits it, then runs `woolly check` and records whether
Woolly flagged the fact as stale. It does this for 50 changes: 30 that make
the fact wrong (renames, deleted files, changed config values, removed
functions) and 20 that don't (comment and layout edits, moved code,
unrelated edits). It reports precision and recall.

Every case runs in its own temporary repo and its own empty Woolly folder.
Nothing touches your real memories, and nothing needs Ollama or a network.

Needs git and Go (it builds woolly from this checkout), or pass --woolly.
"""

import argparse
import json
import os
import statistics
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

# The firmware repo every case starts from.
BASE = {
    "main/pins.h": """#ifndef PINS_H
#define PINS_H

#include "driver/gpio.h"

// SPI bus for the IMU.
#define IMU_SPI_HOST SPI2_HOST
#define IMU_CS GPIO_NUM_5
#define IMU_SCLK GPIO_NUM_18
#define IMU_MISO GPIO_NUM_19
#define IMU_MOSI GPIO_NUM_23

// Motors, front left first, clockwise.
#define MOTOR_FL GPIO_NUM_25
#define MOTOR_FR GPIO_NUM_26
#define MOTOR_BR GPIO_NUM_27
#define MOTOR_BL GPIO_NUM_14

#define LED_STATUS GPIO_NUM_2

#endif
""",
    "main/config.h": """#ifndef CONFIG_H
#define CONFIG_H

#define MOTOR_PWM_HZ 20000
#define MOTOR_PWM_BITS 11
#define LOOP_HZ 1000
#define IMU_SAMPLE_HZ 8000
#define RADIO_BAUD 420000
#define BATTERY_CELLS 4
#define BATTERY_MIN_MV 3300
#define FAILSAFE_MS 250

#endif
""",
    "main/imu.c": """#include "pins.h"
#include "imu.h"

#define IMU_REG_ACCEL 0x3B
#define IMU_REG_PWR 0x6B
#define ACCEL_LSB_PER_G 16384

static spi_device_handle_t imu_dev;

void imu_init(void) {
    spi_bus_config_t bus = {
        .mosi_io_num = IMU_MOSI,
        .miso_io_num = IMU_MISO,
        .sclk_io_num = IMU_SCLK,
    };
    spi_bus_initialize(IMU_SPI_HOST, &bus, SPI_DMA_CH_AUTO);
    imu_reset();
}

void imu_reset(void) {
    imu_write(IMU_REG_PWR, 0x80);
    vTaskDelay(pdMS_TO_TICKS(100));
}

int16_t imu_read_accel(int axis) {
    uint8_t buf[2];
    imu_read(IMU_REG_ACCEL + axis * 2, buf, 2);
    return (int16_t)((buf[0] << 8) | buf[1]);
}

float imu_accel_g(int axis) {
    return imu_read_accel(axis) / (float)ACCEL_LSB_PER_G;
}
""",
    "main/motor.c": """#include "pins.h"
#include "config.h"
#include "driver/ledc.h"

static const int motor_pins[4] = {MOTOR_FL, MOTOR_FR, MOTOR_BR, MOTOR_BL};

void motor_init(void) {
    ledc_timer_config_t timer = {
        .duty_resolution = MOTOR_PWM_BITS,
        .freq_hz = MOTOR_PWM_HZ,
    };
    ledc_timer_config(&timer);
}

void motor_set(int motor, float throttle) {
    if (throttle < 0.0f) throttle = 0.0f;
    if (throttle > 1.0f) throttle = 1.0f;
    uint32_t duty = (uint32_t)(throttle * ((1 << MOTOR_PWM_BITS) - 1));
    ledc_set_duty(LEDC_LOW_SPEED_MODE, motor, duty);
    ledc_update_duty(LEDC_LOW_SPEED_MODE, motor);
}

void motor_stop_all(void) {
    for (int i = 0; i < 4; i++) {
        motor_set(i, 0.0f);
    }
}
""",
    "main/pid.c": """#include "pid.h"

const float PID_KP = 1.2f;
const float PID_KI = 0.05f;
const float PID_KD = 0.02f;
const float PID_I_LIMIT = 0.3f;

float pid_step(pid_t *p, float error, float dt) {
    p->integral += error * dt;
    if (p->integral > PID_I_LIMIT) p->integral = PID_I_LIMIT;
    if (p->integral < -PID_I_LIMIT) p->integral = -PID_I_LIMIT;
    float d = (error - p->last) / dt;
    p->last = error;
    return PID_KP * error + PID_KI * p->integral + PID_KD * d;
}

void pid_reset(pid_t *p) {
    p->integral = 0;
    p->last = 0;
}
""",
    "main/radio.c": """#include "config.h"

typedef enum {
    RADIO_IDLE = 0,
    RADIO_BINDING = 1,
    RADIO_LINKED = 2,
    RADIO_FAILSAFE = 3,
} radio_state_t;

static radio_state_t state = RADIO_IDLE;

void radio_init(void) {
    uart_set_baudrate(UART_NUM_1, RADIO_BAUD);
}

int radio_failsafe_ms(void) {
    return FAILSAFE_MS;
}

void radio_on_packet(const uint8_t *buf, int len) {
    state = RADIO_LINKED;
}
""",
    "main/battery.c": """#include "config.h"

int battery_cell_mv(int pack_mv) {
    return pack_mv / BATTERY_CELLS;
}

bool battery_low(int pack_mv) {
    return battery_cell_mv(pack_mv) < BATTERY_MIN_MV;
}
""",
    "tools/flash.py": """import subprocess

PORT = "/dev/ttyUSB0"
BAUD = 921600


def flash(image):
    subprocess.run(["esptool.py", "--port", PORT, "--baud", str(BAUD), "write_flash", "0x10000", image], check=True)


def erase():
    subprocess.run(["esptool.py", "--port", PORT, "erase_flash"], check=True)
""",
    "CMakeLists.txt": """cmake_minimum_required(VERSION 3.16)
include($ENV{IDF_PATH}/tools/cmake/project.cmake)
project(drone)
""",
}


@dataclass
class Case:
    name: str
    kind: str
    fact: str
    ref: str
    edits: list
    stale: bool  # should Woolly flag it?
    note: str = ""


def rep(path, old, new):
    return ("replace", path, old, new)


def delete(path):
    return ("delete", path)


def move(src, dst):
    return ("move", src, dst)


CASES = [
    # Renames the fact names: should go stale.
    Case("rename imu_read_accel", "rename", "imu_read_accel returns raw counts, not g", "main/imu.c#imu_read_accel",
         [rep("main/imu.c", "imu_read_accel", "imu_read_accel_raw")], True),
    Case("rename imu_reset", "rename", "imu_reset waits 100 ms after the power register write", "main/imu.c#imu_reset",
         [rep("main/imu.c", "imu_reset", "imu_soft_reset")], True),
    Case("rename motor_set", "rename", "motor_set clamps throttle to 0..1 before writing duty", "main/motor.c#motor_set",
         [rep("main/motor.c", "motor_set", "motor_write")], True),
    Case("rename motor_stop_all", "rename", "Call motor_stop_all on failsafe", "main/motor.c#motor_stop_all",
         [rep("main/motor.c", "motor_stop_all", "motors_off")], True),
    Case("rename pid_step", "rename", "pid_step clamps the integral to PID_I_LIMIT", "main/pid.c#pid_step",
         [rep("main/pid.c", "pid_step", "pid_update")], True),
    Case("rename radio_on_packet", "rename", "radio_on_packet is where the link state becomes RADIO_LINKED", "main/radio.c#radio_on_packet",
         [rep("main/radio.c", "radio_on_packet", "radio_handle_frame")], True),
    Case("rename battery_low", "rename", "battery_low compares per-cell millivolts", "main/battery.c#battery_low",
         [rep("main/battery.c", "battery_low", "battery_is_low")], True),
    Case("rename flash in Python", "rename", "tools/flash.py flash() writes the image at 0x10000", "tools/flash.py#flash",
         [rep("tools/flash.py", "def flash(", "def write_image(")], True),

    # Deleted files: should go stale.
    Case("delete battery.c", "delete", "battery_low compares per-cell millivolts", "main/battery.c#battery_low",
         [delete("main/battery.c")], True),
    Case("delete radio.c", "delete", "radio_init sets the UART baud from RADIO_BAUD", "main/radio.c#radio_init",
         [delete("main/radio.c")], True),
    Case("delete flash.py", "delete", "Flash with tools/flash.py", "tools/flash.py",
         [delete("tools/flash.py")], True),
    Case("delete pid.c", "delete", "PID gains live in main/pid.c", "main/pid.c",
         [delete("main/pid.c")], True),
    Case("delete motor.c", "delete", "motor_init sets up the LEDC timer", "main/motor.c#motor_init",
         [delete("main/motor.c")], True),

    # Changed config values: should go stale.
    Case("IMU CS pin", "value", "The IMU chip select is GPIO 5", "main/pins.h#IMU_CS",
         [rep("main/pins.h", "IMU_CS GPIO_NUM_5", "IMU_CS GPIO_NUM_4")], True),
    Case("status LED pin", "value", "The status LED is on GPIO 2", "main/pins.h#LED_STATUS",
         [rep("main/pins.h", "LED_STATUS GPIO_NUM_2", "LED_STATUS GPIO_NUM_13")], True),
    Case("PWM frequency", "value", "Motor PWM runs at 20000 Hz", "main/config.h#MOTOR_PWM_HZ",
         [rep("main/config.h", "MOTOR_PWM_HZ 20000", "MOTOR_PWM_HZ 24000")], True),
    Case("loop rate", "value", "The control loop runs at 1000 Hz", "main/config.h#LOOP_HZ",
         [rep("main/config.h", "LOOP_HZ 1000", "LOOP_HZ 2000")], True),
    Case("radio baud", "value", "The receiver UART runs at 420000 baud", "main/config.h#RADIO_BAUD",
         [rep("main/config.h", "RADIO_BAUD 420000", "RADIO_BAUD 416666")], True),
    Case("battery cells", "value", "The pack is 4 cells", "main/config.h#BATTERY_CELLS",
         [rep("main/config.h", "BATTERY_CELLS 4", "BATTERY_CELLS 6")], True),
    Case("failsafe timeout", "value", "Failsafe trips after 250 ms without packets", "main/config.h#FAILSAFE_MS",
         [rep("main/config.h", "FAILSAFE_MS 250", "FAILSAFE_MS 500")], True),
    Case("PID Kp", "value", "PID_KP is 1.2", "main/pid.c#PID_KP",
         [rep("main/pid.c", "PID_KP = 1.2f", "PID_KP = 0.9f")], True),
    Case("PID Ki", "value", "PID_KI is 0.05", "main/pid.c#PID_KI",
         [rep("main/pid.c", "PID_KI = 0.05f", "PID_KI = 0.08f")], True),
    Case("accel scale", "value", "Accel is 16384 LSB per g", "main/imu.c#ACCEL_LSB_PER_G",
         [rep("main/imu.c", "ACCEL_LSB_PER_G 16384", "ACCEL_LSB_PER_G 8192")], True),
    Case("flash baud in Python", "value", "tools/flash.py flashes at 921600 baud", "tools/flash.py#BAUD",
         [rep("tools/flash.py", "BAUD = 921600", "BAUD = 460800")], True),
    Case("pin map by line range", "value", "Motors are on GPIO 25, 26, 27 and 14", "main/pins.h:14-17",
         [rep("main/pins.h", "MOTOR_BL GPIO_NUM_14", "MOTOR_BL GPIO_NUM_33")], True),
    Case("enum value", "value", "RADIO_FAILSAFE is state 3", "main/radio.c#RADIO_FAILSAFE",
         [rep("main/radio.c", "RADIO_FAILSAFE = 3", "RADIO_FAILSAFE = 4")], True),

    # Removed code: should go stale.
    Case("remove pid_reset", "remove", "Call pid_reset when arming", "main/pid.c#pid_reset",
         [rep("main/pid.c", "void pid_reset(pid_t *p) {\n    p->integral = 0;\n    p->last = 0;\n}\n", "")], True),
    Case("remove erase", "remove", "tools/flash.py erase() wipes the whole flash", "tools/flash.py#erase",
         [rep("tools/flash.py", "\n\ndef erase():\n    subprocess.run([\"esptool.py\", \"--port\", PORT, \"erase_flash\"], check=True)\n", "\n")], True),

    # Hard on purpose: the fact says it differently from the code.
    Case("PWM in kHz", "value", "Motor PWM runs at 20 kHz", "main/config.h#MOTOR_PWM_HZ",
         [rep("main/config.h", "MOTOR_PWM_HZ 20000", "MOTOR_PWM_HZ 24000")], True,
         "The fact says 20 kHz and the code says 20000, so there is no shared value to compare."),
    Case("PWM resolution in words", "value", "Motor PWM has eleven bits of resolution", "main/config.h#MOTOR_PWM_BITS",
         [rep("main/config.h", "MOTOR_PWM_BITS 11", "MOTOR_PWM_BITS 10")], True,
         "The fact spells the number out."),

    # Comment and layout edits: should stay current.
    Case("comment in imu_reset", "comment", "imu_reset waits 100 ms after the power register write", "main/imu.c#imu_reset",
         [rep("main/imu.c", "    vTaskDelay(pdMS_TO_TICKS(100));", "    // The datasheet asks for at least 50 ms.\n    vTaskDelay(pdMS_TO_TICKS(100));")], False),
    Case("comment on IMU_CS", "comment", "The IMU chip select is GPIO 5", "main/pins.h#IMU_CS",
         [rep("main/pins.h", "#define IMU_CS GPIO_NUM_5", "#define IMU_CS GPIO_NUM_5 // moved from 15 on rev A")], False),
    Case("comment in pid_step", "comment", "pid_step clamps the integral to PID_I_LIMIT", "main/pid.c#pid_step",
         [rep("main/pid.c", "    float d = (error - p->last) / dt;", "    /* derivative on error */\n    float d = (error - p->last) / dt;")], False),
    Case("comment in flash.py", "comment", "tools/flash.py flashes at 921600 baud", "tools/flash.py#BAUD",
         [rep("tools/flash.py", "BAUD = 921600", "BAUD = 921600  # the CH340 tops out here")], False),
    Case("reindent motor_set", "layout", "motor_set clamps throttle to 0..1 before writing duty", "main/motor.c#motor_set",
         [rep("main/motor.c", "    if (throttle < 0.0f) throttle = 0.0f;\n    if (throttle > 1.0f) throttle = 1.0f;",
              "\tif (throttle < 0.0f)\n\t\tthrottle = 0.0f;\n\tif (throttle > 1.0f)\n\t\tthrottle = 1.0f;")], False),
    Case("blank lines in config.h", "layout", "Motor PWM runs at 20000 Hz", "main/config.h#MOTOR_PWM_HZ",
         [rep("main/config.h", "#define MOTOR_PWM_HZ 20000\n", "\n#define MOTOR_PWM_HZ   20000\n\n")], False),
    Case("braces on new line", "layout", "battery_low compares per-cell millivolts", "main/battery.c#battery_low",
         [rep("main/battery.c", "bool battery_low(int pack_mv) {", "bool battery_low(int pack_mv)\n{")], False),

    # Edits elsewhere: should stay current.
    Case("new pin elsewhere", "unrelated", "The IMU chip select is GPIO 5", "main/pins.h#IMU_CS",
         [rep("main/pins.h", "#define LED_STATUS GPIO_NUM_2", "#define LED_STATUS GPIO_NUM_2\n#define BUZZER GPIO_NUM_4")], False),
    Case("other config value", "unrelated", "The control loop runs at 1000 Hz", "main/config.h#LOOP_HZ",
         [rep("main/config.h", "FAILSAFE_MS 250", "FAILSAFE_MS 300")], False),
    Case("other function in file", "unrelated", "imu_read_accel returns raw counts, not g", "main/imu.c#imu_read_accel",
         [rep("main/imu.c", "vTaskDelay(pdMS_TO_TICKS(100));", "vTaskDelay(pdMS_TO_TICKS(150));")], False),
    Case("other gain", "unrelated", "PID_KP is 1.2", "main/pid.c#PID_KP",
         [rep("main/pid.c", "PID_KD = 0.02f", "PID_KD = 0.03f")], False),

    # Code that moved: the source should follow it.
    Case("rename not named in fact", "moved", "The accelerometer read takes an axis from 0 to 2", "main/imu.c#imu_read_accel",
         [rep("main/imu.c", "imu_read_accel", "accel_raw")], False),
    Case("rename not named, motor", "moved", "Throttle is clamped to 0..1 before the duty is written", "main/motor.c#motor_set",
         [rep("main/motor.c", "motor_set", "motor_write")], False),
    Case("function to new file", "moved", "Call pid_reset when arming", "main/pid.c#pid_reset",
         [rep("main/pid.c", "void pid_reset(pid_t *p) {\n    p->integral = 0;\n    p->last = 0;\n}\n", ""),
          ("write", "main/pid_util.c", "#include \"pid.h\"\n\nvoid pid_reset(pid_t *p) {\n    p->integral = 0;\n    p->last = 0;\n}\n")], False),
    Case("function to new file, radio", "moved", "radio_init sets the UART baud from RADIO_BAUD", "main/radio.c#radio_init",
         [rep("main/radio.c", "void radio_init(void) {\n    uart_set_baudrate(UART_NUM_1, RADIO_BAUD);\n}\n", ""),
          ("write", "main/radio_uart.c", "#include \"config.h\"\n\nvoid radio_init(void) {\n    uart_set_baudrate(UART_NUM_1, RADIO_BAUD);\n}\n")], False),
    Case("lines added above range", "moved", "Motors are on GPIO 25, 26, 27 and 14", "main/pins.h:14-17",
         [rep("main/pins.h", "// SPI bus for the IMU.", "// Board rev C.\n// SPI bus for the IMU.")], False),
    Case("lines added above range, 2", "moved", "The IMU uses SCLK 18, MISO 19 and MOSI 23", "main/pins.h:9-11",
         [rep("main/pins.h", "#include \"driver/gpio.h\"", "#include \"driver/gpio.h\"\n#include \"sdkconfig.h\"\n")], False),
    Case("file renamed", "moved", "battery_low compares per-cell millivolts", "main/battery.c#battery_low",
         [move("main/battery.c", "main/power.c")], False),
    Case("file renamed, python", "moved", "tools/flash.py flashes at 921600 baud", "tools/flash.py#BAUD",
         [move("tools/flash.py", "tools/esp_flash.py")], False),

    # The code changed but the fact still holds.
    Case("body change keeps values", "changed", "imu_reset waits 100 ms after the power register write", "main/imu.c#imu_reset",
         [rep("main/imu.c", "    imu_write(IMU_REG_PWR, 0x80);", "    imu_write(IMU_REG_PWR, 0x80);\n    imu_flush();")], False),
]


def sh(args, cwd, env=None, check=True):
    return subprocess.run(args, cwd=cwd, env=env, check=check, capture_output=True, text=True)


GIT_ENV = {"GIT_AUTHOR_NAME": "bench", "GIT_AUTHOR_EMAIL": "bench@example.com",
           "GIT_COMMITTER_NAME": "bench", "GIT_COMMITTER_EMAIL": "bench@example.com"}


def apply(repo, edit):
    op = edit[0]
    if op == "replace":
        _, path, old, new = edit
        p = repo / path
        text = p.read_text()
        if old not in text:
            raise SystemExit(f"bench bug: {old!r} not in {path}")
        p.write_text(text.replace(old, new))
    elif op == "delete":
        (repo / edit[1]).unlink()
    elif op == "move":
        sh(["git", "mv", edit[1], edit[2]], repo)
    elif op == "write":
        p = repo / edit[1]
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(edit[2])


def run_case(woolly, case, tmp):
    repo = tmp / "drone"
    home = tmp / "home"
    repo.mkdir()
    for path, text in BASE.items():
        p = repo / path
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(text)
    env = {**os.environ, **GIT_ENV, "WOOLLY_HOME": str(home), "WOOLLY_NO_EMBED": "1", "OLLAMA_HOST": "127.0.0.1:9"}
    sh(["git", "init", "-q", "-b", "main"], repo, env)
    sh(["git", "add", "."], repo, env)
    sh(["git", "commit", "-q", "-m", "base"], repo, env)
    saved = sh([woolly, "remember", case.fact, "--project", "drone", "--source", "agent", "--source-ref", case.ref, "--json"], repo, env)
    mem = json.loads(saved.stdout)["memory"]
    if not mem.get("sources"):
        raise SystemExit(f"bench bug: {case.name}: {case.ref} didn't resolve")
    for edit in case.edits:
        apply(repo, edit)
    sh(["git", "add", "-A"], repo, env)
    sh(["git", "commit", "-q", "-m", case.name], repo, env)
    start = time.perf_counter()
    out = sh([woolly, "check", "--json", "--no-backfill"], repo, env, check=False)
    ms = (time.perf_counter() - start) * 1000
    report = json.loads(out.stdout)
    result = next((r for r in report["results"] if r["id"] == mem["id"]), None)
    verdict = result["verdict"] if result else "unchecked"
    return {"name": case.name, "kind": case.kind, "fact": case.fact, "source": case.ref, "expected": "stale" if case.stale else "current",
            "verdict": verdict, "reason": (result or {}).get("reason", ""), "flagged": verdict == "stale",
            "correct": (verdict == "stale") == case.stale, "check_ms": round(ms, 1), "note": case.note}


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--woolly", help="path to a woolly binary (default: build one from this checkout)")
    ap.add_argument("--json", action="store_true", help="print the full results as JSON")
    args = ap.parse_args()

    with tempfile.TemporaryDirectory() as t:
        woolly = args.woolly
        if not woolly:
            woolly = str(Path(t) / "woolly")
            sh(["go", "build", "-o", woolly, "./cmd/woolly"], ROOT)
        rows = []
        for case in CASES:
            with tempfile.TemporaryDirectory() as ct:
                rows.append(run_case(woolly, case, Path(ct)))

    tp = sum(r["flagged"] and r["expected"] == "stale" for r in rows)
    fp = sum(r["flagged"] and r["expected"] == "current" for r in rows)
    fn = sum(not r["flagged"] and r["expected"] == "stale" for r in rows)
    tn = sum(not r["flagged"] and r["expected"] == "current" for r in rows)
    precision = tp / (tp + fp) if tp + fp else 1.0
    recall = tp / (tp + fn) if tp + fn else 1.0
    summary = {"cases": len(rows), "should_flag": tp + fn, "should_not_flag": fp + tn,
               "true_positives": tp, "false_positives": fp, "false_negatives": fn, "true_negatives": tn,
               "precision": round(precision, 3), "recall": round(recall, 3),
               "median_check_ms": round(statistics.median(r["check_ms"] for r in rows), 1)}

    kinds = {}
    for r in rows:
        k = kinds.setdefault(r["kind"], [0, 0])
        k[0] += r["correct"]
        k[1] += 1

    if args.json:
        print(json.dumps({"summary": summary, "results": rows}, indent=2))
    else:
        print(f"{len(rows)} changes, {tp + fn} that should flag a fact and {fp + tn} that shouldn't.\n")
        print(f"{'kind':<10} correct")
        for kind, (ok, n) in kinds.items():
            print(f"{kind:<10} {ok}/{n}")
        print(f"\nPrecision {precision:.1%}  ({tp} of {tp + fp} flagged facts were really stale)")
        print(f"Recall    {recall:.1%}  ({tp} of {tp + fn} stale facts were flagged)")
        print(f"Median woolly check time {summary['median_check_ms']} ms")
        misses = [r for r in rows if not r["correct"]]
        if misses:
            print("\nWrong calls:")
            for r in misses:
                print(f"  {r['name']}: expected {r['expected']}, got {r['verdict']}. {r['note'] or r['reason']}")

    out_dir = ROOT / "bench" / "stale" / "results"
    out_dir.mkdir(parents=True, exist_ok=True)
    stamp = time.strftime("%Y%m%d-%H%M%S")
    (out_dir / f"{stamp}.json").write_text(json.dumps({"summary": summary, "results": rows}, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
