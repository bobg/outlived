package outlived

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bobg/htree"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
		{
			html: `<div class="fn" style="display:inline">
			         Ishikawa Goemon
			         <br>
			         <style data-mw-deduplicate="TemplateStyles:r886047488">
			           .mw-parser-output .nobold{font-weight:normal}
			         </style>
			         <span class="nobold">石川 五右衛門</span>
			       </div>`,
			want: "Ishikawa Goemon 石川 五右衛門",
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

func TestFindInfoboxImg(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string
	}{
		{
			name: "normal",
			html: `<table class="infobox vcard plainlist" style="width:22em">
			         <tbody>
			           <tr><th colspan="2" style="text-align:center;font-size:125%;font-weight:bold;background-color: #f0e68c"><div style="display:inline;" class="fn">Jerry Garcia</div></th></tr>
			           <tr><td colspan="2" style="text-align:center"><a href="/wiki/File:Jerry-Garcia-01.jpg" class="image"><img alt="Jerry-Garcia-01.jpg" src="//upload.wikimedia.org/wikipedia/commons/thumb/9/97/Jerry-Garcia-01.jpg/227px-Jerry-Garcia-01.jpg" decoding="async" width="227" height="200" srcset="//upload.wikimedia.org/wikipedia/commons/thumb/9/97/Jerry-Garcia-01.jpg/341px-Jerry-Garcia-01.jpg 1.5x, //upload.wikimedia.org/wikipedia/commons/thumb/9/97/Jerry-Garcia-01.jpg/455px-Jerry-Garcia-01.jpg 2x" data-file-width="1091" data-file-height="960"></a><div>Jerry Garcia</div></td></tr>
			           <tr><th colspan="2" style="text-align:center;background-color: #f0e68c">Background information</th></tr>
			           <tr><th scope="row"><span class="nowrap">Birth name</span></th><td class="nickname">Jerome John Garcia</td></tr>
			           <tr><th scope="row">Born</th><td>August 1, 1942<br><a href="/wiki/San_Francisco" title="San Francisco">San Francisco, California</a>, U.S.</td></tr>
			           <tr><th scope="row">Died</th><td>August 9, 1995<span style="display:none">(1995-08-09)</span> (aged&nbsp;53)<br><a href="/wiki/Lagunitas-Forest_Knolls,_California" title="Lagunitas-Forest Knolls, California">Forest Knolls, California</a>, U.S.</td></tr>
			           <tr><th scope="row">Genres</th><td><a href="/wiki/Psychedelic_rock" title="Psychedelic rock">Psychedelic rock</a>, <a href="/wiki/Blues_rock" title="Blues rock">blues rock</a>, <a href="/wiki/Folk_rock" title="Folk rock">folk rock</a>, <a href="/wiki/Country_rock" title="Country rock">country rock</a>, <a href="/wiki/Jam_rock" class="mw-redirect" title="Jam rock">jam rock</a>, <a href="/wiki/Bluegrass_music" title="Bluegrass music">bluegrass</a>, <a href="/wiki/Roots_rock" title="Roots rock">roots rock</a></td></tr>
			           <tr><th scope="row"><span class="nowrap">Occupation(s)</span></th><td class="role">Musician, songwriter</td></tr>
			           <tr><th scope="row">Instruments</th><td class="note"><div class="hlist"><ul><li>Guitar</li><li>pedal steel guitar</li><li>banjo</li><li>vocals</li></ul></div></td></tr>
			           <tr><th scope="row"><span class="nowrap">Years active</span></th><td>1960–1995</td></tr>
			           <tr><th scope="row">Labels</th><td><a href="/wiki/Rhino_Records" class="mw-redirect" title="Rhino Records">Rhino</a>, <a href="/wiki/Arista_Records" title="Arista Records">Arista</a>, <a href="/wiki/Warner_Bros._Records" class="mw-redirect" title="Warner Bros. Records">Warner Bros.</a>, <a href="/wiki/Acoustic_Disc" class="mw-redirect" title="Acoustic Disc">Acoustic Disc</a>, <a href="/wiki/Grateful_Dead_Records" title="Grateful Dead Records">Grateful Dead</a></td></tr>
			           <tr><th scope="row"><span class="nowrap">Associated acts</span></th><td><a href="/wiki/Grateful_Dead" title="Grateful Dead">Grateful Dead</a>, <a href="/wiki/Legion_of_Mary_(band)" title="Legion of Mary (band)">Legion of Mary</a>, <a href="/wiki/Reconstruction_(band)" title="Reconstruction (band)">Reconstruction</a>, <a href="/wiki/Jerry_Garcia_Band" title="Jerry Garcia Band">Jerry Garcia Band</a>, <a href="/wiki/Old_%26_In_the_Way" title="Old &amp; In the Way">Old &amp; In the Way</a>, <a href="/wiki/Jerry_Garcia_Acoustic_Band" title="Jerry Garcia Acoustic Band">Jerry Garcia Acoustic Band</a>, <a href="/wiki/New_Riders_of_the_Purple_Sage" title="New Riders of the Purple Sage">New Riders of the Purple Sage</a>, Hart Valley Drifters, <a href="/wiki/Mother_McCree%27s_Uptown_Jug_Champions" class="mw-redirect" title="Mother McCree's Uptown Jug Champions">Mother McCree's Uptown Jug Champions</a>, <a href="/wiki/Merl_Saunders" title="Merl Saunders">Merl Saunders</a>, Garcia &amp; Grisman, <a href="/wiki/Rainforest_Band" title="Rainforest Band">Rainforest Band</a>, <a href="/wiki/Muruga_Booker" title="Muruga Booker">Muruga Booker</a></td></tr>
			           <tr><th scope="row">Website</th><td><a rel="nofollow" class="external text" href="http://www.jerrygarcia.com">JerryGarcia.com</a><a rel="nofollow" class="external autonumber" href="https://www.jerrygarciamusicarts.com/">[1]</a></td></tr>
			         </tbody>
			       </table>`,
			want: "//upload.wikimedia.org/wikipedia/commons/thumb/9/97/Jerry-Garcia-01.jpg/227px-Jerry-Garcia-01.jpg",
		},
		{
			name: "no image",
			html: `<table class="infobox vcard plainlist" style="width:22em">
			         <tbody>
			           <tr><th colspan="2" style="text-align:center;font-size:125%;font-weight:bold;background-color: #f0e68c"><div style="display:inline;" class="fn">Jerry Garcia</div></th></tr>
			           <tr><th colspan="2" style="text-align:center;background-color: #f0e68c">Background information</th></tr>
			           <tr><th scope="row"><span class="nowrap">Birth name</span></th><td class="nickname">Jerome John Garcia</td></tr>
			           <tr><th scope="row">Born</th><td>August 1, 1942<br><a href="/wiki/San_Francisco" title="San Francisco">San Francisco, California</a>, U.S.</td></tr>
			           <tr><th scope="row">Died</th><td>August 9, 1995<span style="display:none">(1995-08-09)</span> (aged&nbsp;53)<br><a href="/wiki/Lagunitas-Forest_Knolls,_California" title="Lagunitas-Forest Knolls, California">Forest Knolls, California</a>, U.S.</td></tr>
			           <tr><th scope="row">Genres</th><td><a href="/wiki/Psychedelic_rock" title="Psychedelic rock">Psychedelic rock</a>, <a href="/wiki/Blues_rock" title="Blues rock">blues rock</a>, <a href="/wiki/Folk_rock" title="Folk rock">folk rock</a>, <a href="/wiki/Country_rock" title="Country rock">country rock</a>, <a href="/wiki/Jam_rock" class="mw-redirect" title="Jam rock">jam rock</a>, <a href="/wiki/Bluegrass_music" title="Bluegrass music">bluegrass</a>, <a href="/wiki/Roots_rock" title="Roots rock">roots rock</a></td></tr>
			           <tr><th scope="row"><span class="nowrap">Occupation(s)</span></th><td class="role">Musician, songwriter</td></tr>
			           <tr><th scope="row">Instruments</th><td class="note"><div class="hlist"><ul><li>Guitar</li><li>pedal steel guitar</li><li>banjo</li><li>vocals</li></ul></div></td></tr>
			           <tr><th scope="row"><span class="nowrap">Years active</span></th><td>1960–1995</td></tr>
			           <tr><th scope="row">Labels</th><td><a href="/wiki/Rhino_Records" class="mw-redirect" title="Rhino Records">Rhino</a>, <a href="/wiki/Arista_Records" title="Arista Records">Arista</a>, <a href="/wiki/Warner_Bros._Records" class="mw-redirect" title="Warner Bros. Records">Warner Bros.</a>, <a href="/wiki/Acoustic_Disc" class="mw-redirect" title="Acoustic Disc">Acoustic Disc</a>, <a href="/wiki/Grateful_Dead_Records" title="Grateful Dead Records">Grateful Dead</a></td></tr>
			           <tr><th scope="row"><span class="nowrap">Associated acts</span></th><td><a href="/wiki/Grateful_Dead" title="Grateful Dead">Grateful Dead</a>, <a href="/wiki/Legion_of_Mary_(band)" title="Legion of Mary (band)">Legion of Mary</a>, <a href="/wiki/Reconstruction_(band)" title="Reconstruction (band)">Reconstruction</a>, <a href="/wiki/Jerry_Garcia_Band" title="Jerry Garcia Band">Jerry Garcia Band</a>, <a href="/wiki/Old_%26_In_the_Way" title="Old &amp; In the Way">Old &amp; In the Way</a>, <a href="/wiki/Jerry_Garcia_Acoustic_Band" title="Jerry Garcia Acoustic Band">Jerry Garcia Acoustic Band</a>, <a href="/wiki/New_Riders_of_the_Purple_Sage" title="New Riders of the Purple Sage">New Riders of the Purple Sage</a>, Hart Valley Drifters, <a href="/wiki/Mother_McCree%27s_Uptown_Jug_Champions" class="mw-redirect" title="Mother McCree's Uptown Jug Champions">Mother McCree's Uptown Jug Champions</a>, <a href="/wiki/Merl_Saunders" title="Merl Saunders">Merl Saunders</a>, Garcia &amp; Grisman, <a href="/wiki/Rainforest_Band" title="Rainforest Band">Rainforest Band</a>, <a href="/wiki/Muruga_Booker" title="Muruga Booker">Muruga Booker</a></td></tr>
			           <tr><th scope="row">Website</th><td><a rel="nofollow" class="external text" href="http://www.jerrygarcia.com">JerryGarcia.com</a><a rel="nofollow" class="external autonumber" href="https://www.jerrygarciamusicarts.com/">[1]</a></td></tr>
			         </tbody>
			       </table>`,
		},
		{
			name: "image after second TH",
			html: `<table class="infobox vcard" style="width:22em">
			         <caption class="fn">
			           <span class="fn">Paul Hunter</span>
			         </caption>
			         <tbody>
			           <tr>
			             <th scope="row">Born</th>
			             <td>
			               <span style="display:none">
			                 (<span class="bday">1978-10-14</span>)
			               </span>
			               14 October 1978<br>
			               <span class="birthplace">
			                 <span style="font-size:90%;">
			                   <a href="/wiki/Leeds" title="Leeds">Leeds</a>,
			                   <a href="/wiki/West_Yorkshire" title="West Yorkshire">Yorkshire</a>,
			                   <a href="/wiki/England" title="England">England</a>
			                 </span>
			               </span>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row">Died</th>
			             <td>
			               9 October 2006<span style="display:none">(2006-10-09)</span> (aged&nbsp;27)<br>
			               <span class="deathplace">
			                 <span style="font-size:90%;">
			                   <a href="/wiki/Huddersfield" title="Huddersfield">Huddersfield</a>,
			                   <a href="/wiki/West_Yorkshire" title="West Yorkshire">Yorkshire</a>,
			                   <a href="/wiki/England" title="England">England</a>
			                 </span>
			               </span>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row">Sport country</th>
			             <td>
			               <span class="flagicon"><img alt="" src="//upload.wikimedia.org/wikipedia/en/thumb/b/be/Flag_of_England.svg/23px-Flag_of_England.svg.png" decoding="async" width="23" height="14" class="thumbborder" srcset="//upload.wikimedia.org/wikipedia/en/thumb/b/be/Flag_of_England.svg/35px-Flag_of_England.svg.png 1.5x, //upload.wikimedia.org/wikipedia/en/thumb/b/be/Flag_of_England.svg/46px-Flag_of_England.svg.png 2x" data-file-width="800" data-file-height="480">&nbsp;</span>
			               <a href="/wiki/England" title="England">England</a>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row">
			               <a href="/wiki/List_of_snooker_player_nicknames" title="List of snooker player nicknames">Nickname</a>
			             </th>
			             <td>
			               <div class="plainlist nowrap">
			                 <ul>
			                   <li>Beckham of the Baize</li>
			                   <li>The Man with the Golden Cue</li>
			                 </ul>
			               </div>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row">Professional</th>
			             <td>1995–2006</td>
			           </tr>
			           <tr>
			             <th scope="row">
			               <span class="nowrap">
			                 Highest <a href="/wiki/Snooker_world_rankings" title="Snooker world rankings">ranking</a>
			               </span>
			             </th>
			             <td>
			               4
			               <span style="font-size:90%;">
			                 (<a href="/wiki/Snooker_world_rankings_2004/2005" title="Snooker world rankings 2004/2005">2004/2005</a>)
			               </span><sup id="cite_ref-cajt.pwp.blueyonder.co.uk_1-0" class="reference"><a href="#cite_note-cajt.pwp.blueyonder.co.uk-1">[1]</a></sup>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row"><span class="nowrap">Career winnings</span></th>
			             <td><a href="/wiki/Pound_sterling" title="Pound sterling">£</a>1,535,730</td>
			           </tr>
			           <tr>
			             <th scope="row">
			               Highest <dfn id=""><a href="/wiki/Glossary_of_cue_sports_terms#break" title="Glossary of cue sports terms"><span title="See entry at: Glossary of cue sports terms § Break" style="color:inherit;" class="glossary-link">break</span></a></dfn>
			             </th>
			             <td>
			               <b>146</b>: <br><span style="font-size:90%;">2004 Premier League</span>
			             </td>
			           </tr>
			           <tr>
			             <th scope="row">
			               <a href="/wiki/Century_break" title="Century break">Century breaks</a>
			             </th>
			             <td>114</td>
			           </tr>
			           <tr>
			             <th colspan="2" style="text-align:center;background-color: lightgreen">Tournament wins</th>
			           </tr>
			           <tr>
			             <th scope="row"><a href="/wiki/List_of_snooker_players_by_number_of_ranking_titles" title="List of snooker players by number of ranking titles">Ranking</a></th>
			             <td>3</td>
			           </tr>
			           <tr><th scope="row">Non-ranking</th><td>3</td></tr>
			         </tbody>
			       </table>`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(c.html))
			if err != nil {
				t.Fatal(err)
			}
			node = htree.FindEl(node, func(n *html.Node) bool { return n.DataAtom == atom.Table })
			got, _ := findInfoboxImg(node)
			if got != c.want {
				t.Errorf(`got %s, want "%s"`, got, c.want)
			}
		})
	}
}
