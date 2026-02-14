class Mur < Formula
  desc "Invisible continuous learning system for AI coding assistants"
  homepage "https://github.com/mur-run/mur-core"
  url "https://github.com/mur-run/mur-core/archive/refs/tags/v1.7.0.tar.gz"
  sha256 "604c9304cd16058101383063aeec1325242b7816bcb64a01d1e74f144b238dd1"
  license "MIT"
  head "https://github.com/mur-run/mur-core.git", branch: "main"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    ldflags = %W[
      -s -w
      -X github.com/mur-run/mur-core/cmd/mur/cmd.Version=#{version}
      -X github.com/mur-run/mur-core/cmd/mur/cmd.Commit=#{tap.user}
      -X github.com/mur-run/mur-core/cmd/mur/cmd.BuildDate=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags: ldflags.join(" ")), "./cmd/mur"
  end

  def post_install
    ohai "mur installed! Run 'mur init --hooks' to get started."
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/mur version")
  end
end
