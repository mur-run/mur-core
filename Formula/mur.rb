class Mur < Formula
  desc "Invisible continuous learning system for AI coding assistants"
  homepage "https://github.com/mur-run/mur-core"
  url "https://github.com/mur-run/mur-core/archive/refs/tags/v0.9.0.tar.gz"
  sha256 "PLACEHOLDER"  # Will be updated by release workflow
  license "MIT"
  head "https://github.com/mur-run/mur-core.git", branch: "main"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/mur"
  end

  def post_install
    ohai "mur installed! Run 'mur init' to get started."
  end

  test do
    assert_match "mur #{version}", shell_output("#{bin}/mur version")
  end
end
