class Woolly < Formula
  desc "Local project memory for coding agents that flags stale facts"
  homepage "https://usewoolly.com"
  version "0.21.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.21.0/woolly_darwin_arm64.tar.gz"
      sha256 "5435d7a785dbc8a08a527d8d6f9c278358c14583f745368fd2be52a23abb1597"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.21.0/woolly_darwin_amd64.tar.gz"
      sha256 "3de5151ebaaf619643177abcb6bf62a76ad158cb14ff743c870c386d3e6ab2f3"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/rjcuff/woolly/releases/download/v0.21.0/woolly_linux_arm64.tar.gz"
      sha256 "4e944369dd355f7cb8a7d3a25c4e60816d28ea0903ea0fe0347d901a82a0a4ca"
    end
    on_intel do
      url "https://github.com/rjcuff/woolly/releases/download/v0.21.0/woolly_linux_amd64.tar.gz"
      sha256 "e1af73328e732bc6ed9cfbac24c25428d613b99d57d310f74385cfc0761568f1"
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
