class Woolly < Formula
  desc "Local project memory for coding agents that flags stale facts"
  homepage "https://usewoolly.com"
  version "0.20.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.1/woolly_darwin_arm64.tar.gz"
      sha256 "91e80631c2c65da1ab471d4a6c9e9cd25048a16e52dec7f51e4a62b431263b14"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.1/woolly_darwin_amd64.tar.gz"
      sha256 "40c92570cdd396c1b1bd9c8dad190947f48847f09a7a186f6e9ef2b56bfddb43"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.1/woolly_linux_arm64.tar.gz"
      sha256 "772d2b87e7e4cb173961656594cd473e1888a145a9d772e261615ea78c06226a"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.20.1/woolly_linux_amd64.tar.gz"
      sha256 "81d82df40d018ef8fa3bfc89b1b4530218cdc8a25f46b8850f8cc87a7d632687"
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
