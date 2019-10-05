package outlived

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestFindFullName(t *testing.T) {
	cases := []struct {
		html string
		want string
	}{
		{
			html: `<div class="fn">foo</div>`,
			want: "foo",
		},
		{
			html: `<div class="fn"> foo </div>`,
			want: "foo",
		},
		{
			html: `<div class="fn" style="text-align:center;font-size:125%;font-weight:bold">
			         <span class="honorific-prefix" style="font-size: 77%; font-weight: normal;">
			           <a href="/wiki/Sir" title="Sir">Sir</a>
			         </span>
			         <br>
			         <span class="fn">Norman Wisdom</span>
			         <br>
			         <span class="honorific-suffix" style="font-size: 77%; font-weight: normal;">
			           <span class="noexcerpt" style="font-size:100;">
			             <a href="/wiki/Officer_of_the_Order_of_the_British_Empire" class="mw-redirect" title="Officer of the Order of the British Empire">OBE</a>
			           </span>
			         </span>
			       </div>`,
			want: "Sir Norman Wisdom OBE",
		},
		{
			html: `<div class="fn" style="text-align:center;font-size:125%;font-weight:bold;color: black; background-color: #EFEBFF">
			         <small>
			           <a href="/wiki/Her_Grace" class="mw-redirect" title="Her Grace">Her Grace</a>
			         </small>
			         <br>
			         The Duchess of Devonshire
			         <br>
			         <span class="noexcerpt" style="font-size:85%;">
			           <a href="/wiki/Dame_Commander_of_the_Royal_Victorian_Order" class="mw-redirect" title="Dame Commander of the Royal Victorian Order">
			             DCVO
			           </a>
			         </span>
			       </div>`,
			want: "Her Grace The Duchess of Devonshire DCVO",
		},
		{
			html: `<div class="fn" style="text-align:center;font-size:125%;font-weight:bold;background-color: #cbe; font-size: 125%">
			         Charles V
			         <sup id="cite_ref-1" class="reference"><a href="#cite_note-1">[a]</a></sup>
			       </div>`,
			want: "Charles V",
		},
		{
			html: `<div class="fn" style="text-align:center;font-size:125%;font-weight:bold;background-color: #cbe; font-size: 125%">Mehmed IV<br>محمد رابع</div>`,
			want: "Mehmed IV محمد رابع",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%02d", i), func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(c.html))
			if err != nil {
				t.Fatal(err)
			}
			got, err := findFullName(node)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}
