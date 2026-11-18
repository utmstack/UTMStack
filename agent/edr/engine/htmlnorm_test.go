package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Vectors produced by the reference's HTML normalizer (listed in the parity
// research notes, normalization.md section 7): the nocomment and notags
// views, the javascript view and a data: URI file. H43-H45 and J2 are
// checked byte for byte.
func TestHTMLNormalizationVectors(t *testing.T) {
	for _, tt := range []struct {
		name, in, noComment string
		checkPage           bool
		noTags, js, rfc     string
		noScript            bool
	}{
		{"H1", "<HTML><Body BGCOLOR=\"#FFFFFF\">Hello   World!</Body></HTML>", "<html><body bgcolor=\"#ffffff\">hello world!</body></html>", true, " hello world! ", "", "", true},
		{"H2", "a<!-- x -->b", "ab", true, "a b", "", "", false},
		{"H3", "a <!-- x --> b", "a b", true, "a b", "", "", false},
		{"H4", "<!-- a > b -->c", "b -->c", true, " b -->c", "", "", false},
		{"H5", "<p>Caf\xc3\xa9 &amp; cr&egrave;me &AMP; &nbsp;x&lt;y &#65;&#x42;&#0;z &#65b &euro; &yuml;</p>", "<p>caf\xc3\xa9 & cr\xe8me &amp; \xa0x<y ab\x00z Ab &#x20ac; &#x00ff;</p>", true, " caf\xc3\xc3 cr me x y z b ", "", "", false},
		{"H6", "<a HREF='HTTP://Evil.Example/X' title=\"a b\">Click</a>", "<a href=\"http://evil.example/x\" title=\"a b\">click</a>", true, " HTTP://Evil.Example/X click ", "", "", false},
		{"H7", "<img src=Pic.GIF>", "<img src=pic.gif>", true, " ", "", "", false},
		{"H9", "<a href=x> ", "<a href=x>", true, " x ", "", "", false},
		{"H10", "<script>var x = 1;\r\ndocument.write(x);</script>after", "<script>varx=1;document.write(x);</script>after", true, " after", "<script>var n000=1;document.write(n000);</script>", "", false},
		{"H11", "<script language='javascript'>function foo(a, b){return a+b} foo(1,2);</script>", "<script language=\"javascript\">functionfoo(a,b){returna+b}foo(1,2);</script>", true, " ", "<script>function n000(n001,n002){return n001+n002}foo(1,2);</script>", "", false},
		{"H12", "<script>var s = \"A\" + \"B\" + 'c'; eval(\"var q=1;\"); d = unescape(\"%41%u0042\");</script>", "<script>vars=\"a\"+\"b\"+\"c\";eval(\"varq=1;\");d=unescape(\"%41%u0042\");</script>", true, " ", "<script>var n000=\"abc\";var n001=1;d=\"ab\";</script>", "", false},
		{"H13", "<script>// hello world\nx=1</script>", "<script>//helloworldx=1</script>", true, " ", "<script>hello world x=1</script>", "", false},
		{"H14", "<img src=\"data:text/plain;base64,SGVsbG8=\" alt=y>", "<img src=\"data\"alt=y>", true, " ", "", "From html-normalise\nContent-type: text/plain;base64\nContent-transfer-encoding: base64\n\nSGVsbG8=", false},
		{"H15", "<img src=data:,a%41b%zc&amp;d>tail", "<img src=data&tail", true, " ", "", "From html-normalise\nContent-type: \n\naAb%0cd", false},
		{"H16", "<style>body{background:url(data:image/gif;base64,R0lGODlh)}</style>text", "<style>body{background:url(data:image/gif;base64,r0lgodlh)}</style>text", true, " text", "", "", false},
		{"H17", "x&amp", "x&", true, "x ", "", "", false},
		{"H18", "<b>\xa1C\xa1Dz</b>", "<b>..\x00z</b>", true, " ..\x00z ", "", "", false},
		{"H19", "<script>var a=foo(x,y);</script>", "<script>vara=foo(x,y);</script>", true, " ", "<script>var n000=foo(x,n001);</script>", "", false},
		{"H20", "<script>var a=1; var a=2; x=0x1F+1e5+1.5e3+3.14159265;</script>", "<script>vara=1;vara=2;x=0x1f+1e5+1.5e3+3.14159265;</script>", true, " ", "<script>var n001=1;var n001=2;x=0 x1f+1 e5+1500+3.14159;</script>", "", false},
		{"H21", "<a href=\"x>y\">t</a>", "<a href=\"x>y\">t</a>", true, " x>y t ", "", "", false},
		{"H22", "<script>x = \"</b>\"; if (a<b) {}</script>", "<script>x=\"</b>\";if(a<b){}</script>", true, " ", "<script>x=\"</b>\";if(a<b){}</script>", "", false},
		{"H23", "<SCRIPT>alert('Hi')</SCRIPT >z", "<script>alert(\"hi\")</script>z", true, " >z", "<script>alert(\"hi\")</script</script>", "", false},
		{"H24", "<i>&#x41;&#65;&#1601;&#65601;&#128;&#255;&#12290;&lt x&#;y</i> ", "<i>aa&#x0641;A\x80&#x00ff;.&lt x\x00y</i>", true, " x y ", "", "", false},
		{"H25", "<a href=\"&#104;ttp://X/&amp;q\" >go</a> ", "<a href=\"http://x/&q\">go</a>", true, " http://X/&q go ", "", "", false},
		{"H26", "a \xc3\xa9 b", "a \xc3\xa9 b", true, "a \xc3\xc3b", "", "", false},
		{"H27", "<a title=\"x y>z w\">t</a> ", "<a title=\"x y>z w\">t</a>", true, " z w\">t ", "", "", false},
		{"H28", "<p onclick='alert(\"x\")'>Hi</p>", "<p onclick=\"alert(\"x\")\">hi</p>", true, " hi ", "", "", false},
		{"H29", "<br/><BR /><input type=text disabled>", "<br/><br /><input type=text disabled>", true, " ", "", "", false},
		{"H30", "<script src=\"a.js\"></script><script>\nvar  s = 'A\\x41\\u0042\\n';\n</script> <script>b()</script>", "<script src=\"a.js\"></script><script>vars=\"a\\x41\\u0042\\n\";</script><script>b()</script>", true, " ", "<script></script>\n<script>var n000=\"aab \";</script>\n<script>b()</script>", "", false},
		{"H31", "<script>/* c */ x = 1 // c2 d\n y = 2</script>", "<script>/*c*/x=1//c2dy=2</script>", true, " ", "<script>x=1 c2 d y=2</script>", "", false},
		{"H32", "<script>function f(a){var b=a;return b}f(2);</script>", "<script>functionf(a){varb=a;returnb}f(2);</script>", true, " ", "<script>function n000(n001){var n002=n001;return n002}f(2);</script>", "", false},
		{"H33", "<script>document.write(unescape(\"%3Cscript%3Ealert(1)%3C/script%3E\"))</script>", "<script>document.write(unescape(\"%3cscript%3ealert(1)%3c/script%3e\"))</script>", true, " ", "<script>document.write(\"<script>alert(1)</script>\")</script>", "", false},
		{"H34", "<script>eval('x='+'1')</script>", "<script>eval(\"x=\"+\"1\")</script>", true, " ", "<script>x=</script>", "", false},
		{"H35", "<![CDATA[x]]>text<!DOCTYPE html>more", "textmore", true, " text more", "", "", false},
		{"H36", "<script>x=1</script", "<script>x=1</script", true, " ", "<script>x=1</</script>", "", false},
		{"H37", "<p> Hello <b> World </b> !</p>", "<p>hello <b>world </b>!</p>", true, " hello world ! ", "", "", false},
		{"H38", "< b >x</ b >", "<b>x</ b>", true, " x ", "", "", false},
		{"H39", "<form action=x><input name=a></form>\n\n<a\nhref=\"u\"\n>L</a>\n", "<form action=x><input name=a></form><a href=\"u\">l</a>", true, " u l ", "", "", false},
		{"H40", "<img src=\"x.png\"/>tail ", "<img src=\"x.png\" />tail", true, " x.png tail ", "", "", false},
		{"H41", "<a href=\"data:text/html,%3Cscript%3E\">x</a> ", "<a href=\"data\">x</a>", true, " x ", "", "From html-normalise\nContent-type: text/html\n\n<script>", false},
		{"H42", "<body onload=\"eval('x')\">&#x3C;script&#x3E;</body>", "<body onload=\"eval(\"x\")\"><script></body>", true, " script ", "", "", true},
		{"H43", "<script language=\"JScript.Encode\">alert('x')</script><b>Hi  there</b> ", "<script language=\"jscript.encode\"><script language=\"javascript\">alert('x')</script><b>hi  there</b> ", true, " ", "<script>alert(\"x\")</script><b>hi there</b></script>", "", false},
		{"H44", "<html><script language=\"JScript.Encode\">#@~^CwAAAA==C^+.D\x60ECkE#aAMAAA==^#~@</script></html>", "<html><script language=\"jscript.encode\"><script language=\"javascript\">alert(\"hi\")</script></html>", true, " ", "<script>alert(\"hi\")</script>", "", false},
		{"H45", "<%@ LANGUAGE = \"VBScript.Encode\" %>\n<% #@~^CgAAAA==\\ko$K6,J(rDAMAAA==^#~@ %>", "<%@ language=\"vbscript.encode\" %><% //~^cgaaaa==\\ko$k6,j(rdamaaa==^#~@ %>msgbox \"x\" %>", true, " msgbox \"x\" %>", "", "", false},
		{"H46", "\xef\xbb\xbf<html><body>Hi</body></html>", "\xef\xbb\xbf<html><body>hi</body></html>", true, "\xef\xbb\xbb hi ", "", "", false},
		{"H48", "<script>var x = [1,2]; function g(){var y=1,z=2; return y}</script>", "<script>varx=[1,2];functiong(){vary=1,z=2;returny}</script>", true, " ", "<script>var n000=[1,2];function n001(){var n002=1,z=2;return n002}</script>", "", false},
		{"H49", "<script>a.b.c = d.e; x = \"a\\\"b\"; y='it\\'s';</script>", "<script>a.b.c=d.e;x=\"a\\\"b\";y=\"it\\\"s\";</script>", true, " ", "<script>a.b.c=d.e;x=\"a\"b\";y=\"it's\";</script>", "", false},
		{"J1", "<script>eval(\"alert(1)\"); eval(foo);</script>", "", false, "", "<script>alert(1)eval(foo);</script>", "", false},
		{"J2", "<script>x=1;dF('bmfsu%26392%263%3A1');y=2;</script> ", "", false, "", "<script>x=1;y=2;</script>alert(1)</script>", "", false},
		{"J3", "<script>var s='\\u0041\\x42\xc3\xa9';</script>", "", false, "", "<script>var n000=\"ab\xc3\xa9\";</script>", "", false},
	} {
		got := normalizeHTML([]byte(tt.in))
		if got.unsupported != "" {
			t.Errorf("%s: unexpectedly unsupported: %s", tt.name, got.unsupported)
		}
		view, unsupported := normalizeJS(got.scripts)
		js := string(view)
		if tt.name == "J2" && unsupported {
			t.Errorf("%s: dF decoder still incomplete", tt.name)
		}
		if tt.checkPage && (string(got.noComment) != tt.noComment || string(got.noTags) != tt.noTags) {
			t.Errorf("%s: page %q text %q; want %q and %q", tt.name, got.noComment, got.noTags, tt.noComment, tt.noTags)
		}
		if tt.js != "" && js != tt.js {
			t.Errorf("%s: javascript %q; want %q", tt.name, js, tt.js)
		}
		if tt.noScript && len(got.scripts) != 0 {
			t.Errorf("%s: javascript %q; want none", tt.name, js)
		}
		if tt.rfc != "" && (len(got.rfc2397) != 1 || string(got.rfc2397[0].compatibility) != tt.rfc) {
			t.Errorf("%s: data: files %q; want %q", tt.name, got.rfc2397, tt.rfc)
		}
	}
}

// Expected javascript views are from ClamAV 1.5.4 sigtool --html-normalise.
// These harmless strings exercise both unescape passes and append ordering.
func TestDFDecoderOracleCases(t *testing.T) {
	for _, tc := range []struct{ name, in, want string }{
		{"key0", `<script>dF('alert(3);0')</script>`, `<script></script>alert(3);</script>`},
		{"key2", `<script>start();dF('cngtv*4+=2');end();</script>`, `<script>start();end();</script>alert(2);</script>`},
		{"two", `<script>dF('alert(3);0');a=1;dF('cngtv*4+=2');</script>`, `<script>a=1;</script>alert(3);alert(2);</script>`},
		{"empty", `<script>dF('');x=1;</script>`, `<script>df("");x=1;</script>`},
		{"non-string", `<script>dF(123);x=1;</script>`, `<script>df(123);x=1;</script>`},
		{"double-quote", `<script>dF("bmfsu%26392%263%3A1");</script>`, `<script></script>alert(1)</script>`},
		{"nested-eval", `<script>eval("dF('bmfsu%26392%263%3A1');")</script>`, `<script>df("bmfsu%26392%263%3a1");</script>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := normalizeHTML([]byte(tc.in))
			got, unsupported := normalizeJS(v.scripts)
			if unsupported || string(got) != tc.want {
				t.Fatalf("javascript %q, incomplete %t; want %q", got, unsupported, tc.want)
			}
		})
	}
}

func TestDFDecoderGeneratedReferenceVectors(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "df_oracle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct{ Name, Source, Output string }
	if err := json.Unmarshal(b, &vectors); err != nil || len(vectors) != 32 {
		t.Fatalf("invalid reference vectors: %d, %v", len(vectors), err)
	}
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			view := normalizeHTML([]byte(v.Source))
			got, unknown := normalizeJS(view.scripts)
			if unknown || string(got) != v.Output {
				t.Fatalf("unknown=%t javascript=%q; want %q", unknown, got, v.Output)
			}
		})
	}
}

// Outputs of the reference's javascript view for unbalanced brackets: the
// counts of '[', '{' and '(' never go below zero, so a stray closer does not
// stop a later ',' from declaring the next name, a stray '}' in a function
// ends the function, and a '(' after a stray ')' is still closed at the end.
func TestJSBracketCountsStopAtZero(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"var a = ], b; b", "<script>var n000=],n001;n001</script>"},
		{"]] var a = 1, b; b", "<script>]]var n000=1,n001;n001</script>"},
		{"a) b(", "<script>a)b(</script>)</script>"},
		{"a)) b((", "<script>a))b((</script>))</script>"},
		{"var a = }, b; b", "<script>var n000=},n001;n001</script>"},
		{"} var a = 1, b; b", "<script>}var n000=1,n001;n001</script>"},
		{"function f(a) } a; var c = 1, d; d", "<script>function n000(n001)}a;var n002=1,n003;n003</script>"},
		{"function f(a) { } } a; var c = 1, d; d", "<script>function n000(n001){}}a;var n002=1,n003;n003</script>"},
		{"function f(a) { ] var b = 1, c; c }", "<script>function n000(n001){]var n002=1,c;c}</script>"},
		{"var a = [}, b; b", "<script>var n000=[},b;b</script>"},
		{"x = ] [ , y; var a = 1, b; b", "<script>x=][,y;var n000=1,n001;n001</script>"},
	} {
		v := normalizeHTML([]byte("<html><script>" + tt.in + "</script>"))
		if got, _ := normalizeJS(v.scripts); string(got) != tt.want {
			t.Errorf("%q: javascript %q; want %q", tt.in, got, tt.want)
		}
	}
}

// Outputs of the reference's javascript view for malformed escapes and large
// numbers: a non-hex escape digit counts as 0 unless it is the last one,
// which sets every bit; code points are written as three bytes even for
// surrogates; a number keeps the low 32 bits of the C conversion, which stops
// at its largest value.
func TestJSEscapesAndNumbersAsReference(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{`x="\u....";`, "x=\"\xef\xbf\xbf\";"},
		{`x="\uZ123";`, "x=\"\xc4\xa3\";"},
		{`x="\u1Z23";`, "x=\"\xe1\x80\xa3\";"},
		{`x="\u12Z3";`, "x=\"\xe1\x88\x83\";"},
		{`x="\u123Z";`, "x=\"\xef\xbf\xbf\";"},
		{`x="\uZ12Z";`, "x=\"\xef\xbf\xbf\";"},
		{`x="\uZZZ1";`, "x=\"\x01\";"},
		{`x="\u1ZZ1";`, "x=\"\xe1\x80\x81\";"},
		{`x="\u00Zf";`, "x=\"\x0f\";"},
		{`x="\uZ000";`, "x=\"\x01\";"},
		{`x="\u0000";`, "x=\"\x01\";"},
		{`x="\uD800";`, "x=\"\xed\xa0\x80\";"},
		{`x="￾";`, "x=\"\xef\xbf\xbe\";"},
		{`x="\u0080";`, "x=\"\xc2\x80\";"},
		{`x="\u12";`, `x="\12";`},
		{`x="\xZ1";`, "x=\"\x01\";"},
		{`x="\x1Z";`, "x=\"\xff\";"},
		{`x="\xZZ";`, "x=\"\xff\";"},
		{`x="\xZ0";`, "x=\"\x01\";"},
		{`x=unescape("%uD800a%uDFFF");`, "x=\"\xed\xa0\x80a\xed\xbf\xbf\";"},
		{"x=2147483648;", "x=-2147483648;"},
		{"x=4294967295;", "x=-1;"},
		{"x=4294967296;", "x=0;"},
		{"x=12345678901;", "x=-539222987;"},
		{"x=99999999999999999999;", "x=-1;"},
		{"x=-2147483649;", "x=--2147483647;"},
		{"x=9223372036854775808;", "x=-1;"},
		{"x=3000000000;", "x=-1294967296;"},
		{"x=00012;", "x=12;"},
		{"x=1.5e400;", "x=inf;"},
	} {
		v := normalizeHTML([]byte("<html><script>" + tt.in + "</script>"))
		if got, want := jsView(v.scripts), "<script>"+tt.want+"</script>"; got != want {
			t.Errorf("%q: javascript %q; want %q", tt.in, got, want)
		}
	}
}

func jsView(scripts [][][]byte) string {
	view, _ := normalizeJS(scripts)
	return string(view)
}

// Outputs of the reference's javascript view for eval() and unescape(),
// learned from probe pages (see normalizeJS and the file comment of
// jsnorm.go). Each page is "<html><script>" + in + "</script>" unless eof
// is set (then the page ends there).
func TestJSEvalAsReference(t *testing.T) {
	for _, tt := range []struct {
		in, want string
		eof      bool
	}{
		// The string's unfinished last token is dropped, or joins the next
		// decoded string.
		{in: `eval("a b")`, want: "a"},
		{in: `eval("a b ")`, want: "a b"},
		{in: `eval("a b")eval("c d")eval("e f")`, want: "a bc de"},
		{in: `eval("a b")x;eval("c d")`, want: "a x;bc"},
		// The pass skips the token after a string that gave none.
		{in: `eval("a")eval("b")x`, want: `eval("b")x`},
		{in: `eval("a")eval("b")eval("c")eval("d")`, want: `eval("b")eval("d")`},
		{in: `eval("a")+eval("b")+eval("c")`, want: "++"},
		{in: `eval("eval('a')")`, want: `eval("a")`},
		{in: `eval("x eval('a b') y")`, want: "x ya"},
		// A ';' after the call goes; unescape() of decoded text stays.
		{in: `x=1; eval("y=2"); z=3`, want: "x=1;y=z=3"},
		{in: `eval("unescape('%41') ")`, want: `unescape("%41")`},
		{in: `eval("'a'+'b' ")`, want: `"ab"`},
		{in: `eval("'x'")+"y"`, want: `"x"+"y"`},
		{in: `eval(unescape("%61 "))`, want: "a"},
		// The block's own reader state reaches the strings.
		{in: `eval("x ")/*`, want: ""},
		{in: `eval("x ")y/*`, want: "y"},
		{in: `eval("x ")//`, want: "x"},
		{in: `<html><script>eval("a b") x`, want: "xa", eof: true},
		{in: `<html><script>eval("var q ")q`, want: "qvar q", eof: true},
		{in: `<html><script>eval("a b") x</scrip`, want: "scripa x</", eof: true},
		// Decoded tokens declare last, into the top-level scope, with a
		// state of their own, and look names up from the eval's scope.
		{in: `eval("var a = 1")var c, d; d`, want: "var n002=var n000,n001;n001"},
		{in: `eval("var ")b;eval("c ")`, want: "var b;n000"},
		{in: `function f(){ var c; eval("c ") }`, want: "function n000(){var n001;n001}"},
		{in: `function f(){ var a; eval("var a ") } a`, want: "function n000(){var n001;var n001}n002"},
		{in: `function g eval("a b") x }`, want: "function n000 a n001}"},
		// After "function" or "function name", at most 8 tokens may follow.
		{in: `function f(){}eval("a b") y y y y`, want: "function n000(){}a y y y y"},
		{in: `function f(){}eval("a b") y y y y y`, want: `function n000(){}eval("a b")y y y y y`},
		{in: `function f(){}x eval("a b") y y y y y y`, want: `function n000(){}x eval("a b")y y y y y y`},
		{in: `function f(){}x x eval("a b") y y y y y y`, want: "function n000(){}x x a y y y y y y"},
		{in: `function f(){};eval("a b") y y y y y y`, want: "function n000(){};a y y y y y y"},
		{in: `function eval("a b") }`, want: "function n001}"},
		{in: `function g eval("a b") function f(){}`, want: `function n000 eval("a b")function n001(){}</n002>`},
		// unescape() folds when any next token arrives, consuming it.
		{in: `unescape("%41"x)`, want: `"a")`},
		{in: `unescape("%41","x")`, want: `"a""x")`},
		{in: `unescape("%41")+"b"`, want: `"a"+"b"`},
		{in: `unescape("%41"+"%42")`, want: `"ab"`},
		{in: `unescape(unescape("%2541"))`, want: `unescape("%41")`},
		{in: `x.unescape("%41")`, want: `x.unescape("%41")`},
		// Numbers: further exponents after a digit; an integer keeps its
		// leading digits.
		{in: "x=1.e5e5;", want: "x=100000;"},
		{in: "x=1.e+e;", want: "x=1;"},
		{in: "x=1.ee;", want: "x=1 e;"},
		{in: "x=1.5e5.5;", want: "x=150000.5;"},
	} {
		page := "<html><script>" + tt.in + "</script>"
		if tt.eof {
			page = tt.in
		}
		v := normalizeHTML([]byte(page))
		if got, want := jsView(v.scripts), "<script>"+tt.want+"</script>"; got != want {
			t.Errorf("%q: javascript %q; want %q", tt.in, got, want)
		}
	}
}

// These exact carried-state outputs were measured with pinned native sigtool.
func TestJSCarriedDecodingNativeExamples(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{`eval("a b")eval("c d")`, `<script>a bc</script>`},
		{`eval("x 'y")eval("z ")`, `<script>x</script>`},
		{`eval("a ")eval("b ")`, `<script>a b</script>`},
	} {
		v := normalizeHTML([]byte("<html><script>" + tc.source + "</script>"))
		got, unsupported := normalizeJS(v.scripts)
		if unsupported || string(got) != tc.want {
			t.Fatalf("%q => %q incomplete=%t want=%q", tc.source, got, unsupported, tc.want)
		}
	}
}
