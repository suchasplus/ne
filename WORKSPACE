load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "33acc4ae0f70502db4b893c9fc1dd7a9bf998c23e7ff2c4517741d4049a976f8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.48.0/rules_go-v0.48.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.48.0/rules_go-v0.48.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "d76bf7a60fd8b050444090dfa2837a4eaf9829e1165618ee35dceca5cbdf58d5",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.37.0/bazel-gazelle-v0.37.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.37.0/bazel-gazelle-v0.37.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_repository(
    name = "com_github_agnivade_levenshtein",
    importpath = "github.com/agnivade/levenshtein",
    sum = "h1:EHBY3UOn1gwdy/VbFwgo4cxecRznFk7fKWN1KOX7eoM=",
    version = "v1.2.1",
)

go_repository(
    name = "com_github_arbovm_levenshtein",
    importpath = "github.com/arbovm/levenshtein",
    sum = "h1:jfIu9sQUG6Ig+0+Ap1h4unLjW6YQJpKZVmUzxsD4E/Q=",
    version = "v0.0.0-20160628152529-48b4e1c0c4d0",
)

go_repository(
    name = "com_github_aymanbagabas_go_osc52_v2",
    importpath = "github.com/aymanbagabas/go-osc52/v2",
    sum = "h1:HwpRHbFMcZLEVr42D4p7XBqjyuxQH5SMiErDT4WkJ2k=",
    version = "v2.0.1",
)

go_repository(
    name = "com_github_aymanbagabas_go_udiff",
    importpath = "github.com/aymanbagabas/go-udiff",
    sum = "h1:TK0fH4MteXUDspT88n8CKzvK0X9O2xu9yQjWpi6yML8=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_bits_and_blooms_bitset",
    importpath = "github.com/bits-and-blooms/bitset",
    sum = "h1:Tquv9S8+SGaS3EhyA+up3FXzmkhxPGjQQCkcs2uw7w4=",
    version = "v1.22.0",
)

go_repository(
    name = "com_github_charmbracelet_colorprofile",
    importpath = "github.com/charmbracelet/colorprofile",
    sum = "h1:k8dTHMd7fgw4bnFd7jXTLZrSU/CQrKnL3m+AxCzDz40=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_charmbracelet_lipgloss",
    importpath = "github.com/charmbracelet/lipgloss",
    sum = "h1:vYXsiLHVkK7fp74RkV7b2kq9+zDLoEU4MZoFqR/noCY=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_charmbracelet_x_ansi",
    importpath = "github.com/charmbracelet/x/ansi",
    sum = "h1:92AGsQmNTRMzuzHEYfCdjQeUzTrgE1vfO5/7fEVoXdY=",
    version = "v0.9.2",
)

go_repository(
    name = "com_github_charmbracelet_x_cellbuf",
    importpath = "github.com/charmbracelet/x/cellbuf",
    sum = "h1:/KBBKHuVRbq1lYx5BzEHBAFBP8VcQzJejZ/IA3iR28k=",
    version = "v0.0.13",
)

go_repository(
    name = "com_github_charmbracelet_x_exp_golden",
    importpath = "github.com/charmbracelet/x/exp/golden",
    sum = "h1:G99klV19u0QnhiizODirwVksQB91TJKV/UaTnACcG30=",
    version = "v0.0.0-20240806155701-69247e0abc2a",
)

go_repository(
    name = "com_github_charmbracelet_x_term",
    importpath = "github.com/charmbracelet/x/term",
    sum = "h1:AQeHeLZ1OqSXhrAWpYUtZyX1T3zVxfpZuEQMIQaGIAQ=",
    version = "v0.2.1",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_dgryski_trifles",
    importpath = "github.com/dgryski/trifles",
    sum = "h1:SG7nF6SRlWhcT7cNTs5R6Hk4V2lcmLz2NsG2VnInyNo=",
    version = "v0.0.0-20230903005119-f50d829f2e54",
)

go_repository(
    name = "com_github_edsrzf_mmap_go",
    importpath = "github.com/edsrzf/mmap-go",
    sum = "h1:6EUwBLQ/Mcr1EYLE4Tn1VdW1A4ckqCQWZBw8Hr0kjpQ=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    importpath = "github.com/inconshreveable/mousetrap",
    sum = "h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_lucasb_eyer_go_colorful",
    importpath = "github.com/lucasb-eyer/go-colorful",
    sum = "h1:1nnpGOrhyZZuNyfu1QjKiUICQ74+3FNCN69Aj6K7nkY=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    sum = "h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=",
    version = "v0.0.20",
)

go_repository(
    name = "com_github_mattn_go_runewidth",
    importpath = "github.com/mattn/go-runewidth",
    sum = "h1:E5ScNMtiwvlvB5paMFdw9p4kSQzbXFikJ5SQO6TULQc=",
    version = "v0.0.16",
)

go_repository(
    name = "com_github_muesli_termenv",
    importpath = "github.com/muesli/termenv",
    sum = "h1:S5AlUN9dENB57rsbnkPyfdGuWIlkmzJjbFf0Tf5FWUc=",
    version = "v0.16.0",
)

go_repository(
    name = "com_github_olekukonko_tablewriter",
    importpath = "github.com/olekukonko/tablewriter",
    sum = "h1:P2Ga83D34wi1o9J6Wh1mRuqd4mF/x/lgBS7N7AbDhec=",
    version = "v0.0.5",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_rivo_uniseg",
    importpath = "github.com/rivo/uniseg",
    sum = "h1:WUdvkW8uEhrYfLC4ZzdpI2ztxP1I582+49Oc5Mq64VQ=",
    version = "v0.4.7",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    sum = "h1:e5/vxKd/rZsfSJMUX1agtjeTDf+qv1/JdBF8gg5k9ZM=",
    version = "v1.8.1",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    sum = "h1:jFzHGLGAlb3ruxLB8MhbI6A8+AQX/2eW4qeyNZXNp2o=",
    version = "v1.0.6",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    sum = "h1:Xv5erBjTwe/5IxqUQTdXv5kgmIvbHo3QQyRwhJsOfJA=",
    version = "v1.10.0",
)

go_repository(
    name = "com_github_urfave_cli_v3",
    importpath = "github.com/urfave/cli/v3",
    sum = "h1:BYFVnhhZ8RqT38DxEYVFPPmGFTEf7tJwySTXsVRrS/o=",
    version = "v3.3.2",
)

go_repository(
    name = "com_github_xo_terminfo",
    importpath = "github.com/xo/terminfo",
    sum = "h1:JVG44RsyaB9T2KIHavMF/ppJZNG9ZpyihvCd0w101no=",
    version = "v0.0.0-20220910002029-abceb7e1c41e",
)

go_repository(
    name = "in_gopkg_yaml_v3",
    importpath = "gopkg.in/yaml.v3",
    sum = "h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=",
    version = "v3.0.1",
)

go_repository(
    name = "io_etcd_go_bbolt",
    importpath = "go.etcd.io/bbolt",
    sum = "h1:TU77id3TnN/zKr7CO/uk+fBCwF2jGcMuw2B/FMAzYIk=",
    version = "v1.4.0",
)

go_repository(
    name = "io_etcd_go_gofail",
    importpath = "go.etcd.io/gofail",
    sum = "h1:p19drv16FKK345a09a1iubchlw/vmRuksmRzgBIGjcA=",
    version = "v0.2.0",
)

go_repository(
    name = "org_golang_x_exp",
    importpath = "golang.org/x/exp",
    sum = "h1:MDc5xs78ZrZr3HMQugiXOAkSZtfTpbJLDr/lwfgO53E=",
    version = "v0.0.0-20220909182711-5c715a9e8561",
)

go_repository(
    name = "org_golang_x_sync",
    importpath = "golang.org/x/sync",
    sum = "h1:3NQrjDixjgGwUOCaF8w2+VYHv0Ve/vGYSbdkTa98gmQ=",
    version = "v0.10.0",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:q3i8TbbEz+JRD9ywIRlyRAQbM0qF7hu24q3teo2hbuw=",
    version = "v0.33.0",
)

go_repository(
    name = "org_uber_go_goleak",
    importpath = "go.uber.org/goleak",
    sum = "h1:2K3zAYmnTNqV73imy9J1T3WC+gmCePx2hEGkimedGto=",
    version = "v1.3.0",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    sum = "h1:S0h4aNzvfcFsC3dRF1jLoaov7oRaKqRGC/pUEJ2yvPQ=",
    version = "v1.10.0",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
    sum = "h1:aJMhYGrd5QSmlpLMr2MftRKl7t8J8PTZPA732ud/XR8=",
    version = "v1.27.0",
)

go_rules_dependencies()

go_register_toolchains(version = "1.24.2")

gazelle_dependencies()
