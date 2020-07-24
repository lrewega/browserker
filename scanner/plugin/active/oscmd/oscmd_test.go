package oscmd_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/plugin/active/oscmd"
	"gitlab.com/browserker/scanner/plugin/plugintest"
)

var leaser = browser.NewLocalLeaser()

func init() {
	leaser.SetHeadless()
	//leaser.SetProxy("http://127.0.0.1:8080")
}

const passwdContents = `root:x:0:0:root:/root:/bin/bash
bin:x:1:1:bin:/bin:/sbin/nologin
daemon:x:2:2:daemon:/sbin:/sbin/nologin
adm:x:3:4:adm:/var/adm:/sbin/nologin
lp:x:4:7:lp:/var/spool/lpd:/sbin/nologin
sync:x:5:0:sync:/sbin:/bin/sync
shutdown:x:6:0:shutdown:/sbin:/sbin/shutdown
halt:x:7:0:halt:/sbin:/sbin/halt
mail:x:8:12:mail:/var/spool/mail:/sbin/nologin
news:x:9:13:news:/etc/news:
uucp:x:10:14:uucp:/var/spool/uucp:/sbin/nologin
operator:x:11:0:operator:/root:/sbin/nologin
games:x:12:100:games:/usr/games:/sbin/nologin
gopher:x:13:30:gopher:/var/gopher:/sbin/nologin
ftp:x:14:50:FTP User:/var/ftp:/sbin/nologin
nobody:x:99:99:Nobody:/:/sbin/nologin
nscd:x:28:28:NSCD Daemon:/:/sbin/nologin
vcsa:x:69:69:virtual console memory owner:/dev:/sbin/nologin
ntp:x:38:38::/etc/ntp:/sbin/nologin
pcap:x:77:77::/var/arpwatch:/sbin/nologin
dbus:x:81:81:System message bus:/:/sbin/nologin
avahi:x:70:70:Avahi daemon:/:/sbin/nologin
rpc:x:32:32:Portmapper RPC user:/:/sbin/nologin
mailnull:x:47:47::/var/spool/mqueue:/sbin/nologin
smmsp:x:51:51::/var/spool/mqueue:/sbin/nologin
apache:x:48:48:Apache:/var/www:/sbin/nologin
sshd:x:74:74:Privilege-separated SSH:/var/empty/sshd:/sbin/nologin
dovecot:x:97:97:dovecot:/usr/libexec/dovecot:/sbin/nologin
oprofile:x:16:16:Special user account to be used by OProfile:/home/oprofile:/sbin/nologin
rpcuser:x:29:29:RPC Service User:/var/lib/nfs:/sbin/nologin
nfsnobody:x:65534:65534:Anonymous NFS User:/var/lib/nfs:/sbin/nologin
xfs:x:43:43:X Font Server:/etc/X11/fs:/sbin/nologin
haldaemon:x:68:68:HAL daemon:/:/sbin/nologin
avahi-autoipd:x:100:156:avahi-autoipd:/var/lib/avahi-autoipd:/sbin/nologin
gdm:x:42:42::/var/gdm:/sbin/nologin`

func TestOSCMD(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()
	called := false

	toTest := [...]plugintest.AttackTests{
		{
			FormHandler: func(c *gin.Context) {
				user := c.PostForm("username")
				resp := "<html><body>You made it!</body></html>"
				if user == ";cat /etc/passwd" {
					resp = passwdContents
					called = true
				}

				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			URL: "http://localhost:%s/forms/simplePOST.html",
		},
		{
			FormHandler: func(c *gin.Context) {
				user, _ := c.GetQuery("username")
				resp := "<html><body>You made it!</body></html>"
				if user == "|cat /etc/passwd" {
					resp = passwdContents
					called = true
				}

				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			URL: "http://localhost:%s/forms/simpleGET.html",
		},
	}

	for _, attackTest := range toTest {
		p, srv := plugintest.RunTestServer("../../plugintest/testdata/forms", "/form/result", attackTest.FormHandler)
		defer srv.Shutdown(ctx)

		target := fmt.Sprintf(attackTest.URL, p)
		targetURL, _ := url.Parse(target)
		ctx := context.Background()
		bCtx := mock.MakeMockContext(ctx, targetURL)
		bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)
		bCtx.Scope = scanner.NewScopeService(targetURL)
		bCtx.Reporter = mock.MakeMockReporter()

		browser, port, navResults, err := plugintest.GetNewNavPaths(bCtx, pool, target)
		if err != nil {
			t.Fatalf("error getting new nav paths: %s\n", err)
		}

		if len(navResults) == 0 {
			t.Fatal("did not find form nav action")
		}

		bCtx.PluginServicer = mock.MakeMockPluginServicer()
		bCtx.PluginServicer.Register(oscmd.New(bCtx.PluginServicer))
		t.Logf("Attacking With Plugin")
		plugintest.AttackWithPlugin(bCtx, browser, navResults)
		if !called {
			t.Fatalf("attack was not successful: %s\n", target)
		}

		called = false
		pool.Return(ctx, port)
		srv.Shutdown(ctx)
	}
}
