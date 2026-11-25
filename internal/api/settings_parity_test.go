package api

import (
	"encoding/json"
	"testing"

	"github.com/overnite22/relay-public/internal/store"
)

// Réglages partagés avec Relay Desktop (core/src/conn.rs) → paramètres guacd.
func TestBuildConfigDesktopSettings(t *testing.T) {
	rdp := func(host, settings string) map[string]string {
		t.Helper()
		cfg, err := buildConfig(store.Entry{Kind: "rdp", Host: host, Port: 3389, Settings: json.RawMessage(settings)},
			secrets{username: "u", domain: "D", password: "pw"}, 1280, 720, 96, "", "")
		if err != nil {
			t.Fatal(err)
		}
		p := cfg.Params
		p["_size"] = itoa(cfg.Width) + "x" + itoa(cfg.Height)
		return p
	}
	p := rdp("pc.example.test", `{}`)
	if p["_size"] != "1280x720" || p["resize-method"] != "display-update" || p["disable-audio"] != "" ||
		p["enable-printing"] != "" || p["gateway-hostname"] != "" {
		t.Fatalf("défauts : %v", p)
	}
	p = rdp("pc.example.test", `{"resolution":"fixed","width":1600,"height":900}`)
	if p["_size"] != "1600x900" || p["resize-method"] != "" {
		t.Fatalf("taille fixe : %v", p)
	}
	if p = rdp("h", `{"resolution":"fixed","width":50,"height":900}`); p["_size"] != "1280x720" || p["resize-method"] == "" {
		t.Fatalf("taille fixe invalide : %v", p)
	}
	if p = rdp("h", `{"audioMode":"remote"}`); p["disable-audio"] != "true" {
		t.Fatalf("son distant : %v", p)
	}
	if p = rdp("h", `{"disableAudio":true}`); p["disable-audio"] != "true" {
		t.Fatalf("son désactivé : %v", p)
	}
	if p = rdp("h", `{"redirectPrinters":true}`); p["enable-printing"] != "true" || p["printer-name"] != "Relay" {
		t.Fatalf("imprimantes : %v", p)
	}
	p = rdp("10.0.0.5", `{"gatewayUsage":"always","gatewayHost":"gw.example.test:8443"}`)
	if p["gateway-hostname"] != "gw.example.test" || p["gateway-port"] != "8443" ||
		p["gateway-username"] != "u" || p["gateway-password"] != "pw" || p["gateway-domain"] != "D" {
		t.Fatalf("passerelle : %v", p)
	}
	p = rdp("pc.example.test", `{"gatewayUsage":"always","gatewayHost":"gw.example.test","gatewaySameCreds":false}`)
	if p["gateway-hostname"] != "gw.example.test" || p["gateway-port"] != "" || p["gateway-username"] != "" {
		t.Fatalf("passerelle sans identifiants : %v", p)
	}
	for host, want := range map[string]string{"10.0.0.5": "", "srv": "", "pc.example.test": "gw"} {
		if p = rdp(host, `{"gatewayUsage":"auto","gatewayHost":"gw"}`); p["gateway-hostname"] != want {
			t.Fatalf("passerelle auto %s : %v", host, p)
		}
	}
	if p = rdp("pc.example.test", `{"gatewayHost":"gw"}`); p["gateway-hostname"] != "" {
		t.Fatalf("passerelle sans usage : %v", p)
	}

	ssh := func(settings, lang string) map[string]string {
		cfg, _ := buildConfig(store.Entry{Kind: "ssh", Host: "h", Port: 22, Settings: json.RawMessage(settings)}, secrets{}, 0, 0, 0, "", lang)
		return cfg.Params
	}
	if p = ssh(`{"charset":"cp850"}`, "fr-FR"); p["locale"] != "" || p["server-charset"] != "" {
		t.Fatalf("charset seul : %v", p)
	}
	for lang, want := range map[string]string{"fr-FR": "fr_FR.UTF-8", "de": "de.UTF-8", "": "en_US.UTF-8", "x;y": "en_US.UTF-8"} {
		if p = ssh(`{"sendLocale":true}`, lang); p["locale"] != want {
			t.Fatalf("locale %q : %v", lang, p["locale"])
		}
	}
}

func itoa(v int) string { b, _ := json.Marshal(v); return string(b) }
