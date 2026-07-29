package engine

import (
	"strings"
	"testing"

	"github.com/goed2k/core"
	"github.com/goed2k/core/protocol"
)

func TestEd2kLinkForAddPreservesAICH(t *testing.T) {
	root, err := protocol.AICHHashFromString("A9993E364706816ABA3E25717850C26C9CD0D89D")
	if err != nil {
		t.Fatal(err)
	}
	original := "ed2k://|file|demo.bin|1024|31D6CFE0D16AE931B73C59D7E0C089C0|h=" + root.Base32() + "|/"
	link, err := goed2k.ParseEMuleLink(original)
	if err != nil {
		t.Fatal(err)
	}
	got := ed2kLinkForAdd(link, "", original)
	if got != original {
		t.Fatalf("got %q, want original link with AICH preserved", got)
	}
}

func TestEd2kLinkForAddRenamedDropsExtensions(t *testing.T) {
	hash, err := protocol.HashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	if err != nil {
		t.Fatal(err)
	}
	original := "ed2k://|file|demo.bin|1024|31D6CFE0D16AE931B73C59D7E0C089C0|h=ABCDEF0123456789ABCDEF0123456789ABCDEF|/"
	link := goed2k.EMuleLink{
		Hash:        hash,
		NumberValue: 1024,
		StringValue: "demo.bin",
		Type:        goed2k.LinkFile,
	}
	got := ed2kLinkForAdd(link, "renamed.bin", original)
	want := goed2k.FormatLink("renamed.bin", 1024, hash)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if strings.Contains(got, "h=") {
		t.Fatalf("renamed link should not keep AICH segment: %q", got)
	}
}
