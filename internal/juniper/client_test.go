package juniper

import ("testing"; "net/netip"; "strings")

func TestXMLActiveRoute(t *testing.T) {
 s := `<rpc-reply><route-information><route-table><rt><rt-destination>192.0.2.0</rt-destination><rt-prefix-length>24</rt-prefix-length><tsi>Localpref: 105 AS path: 64512 I</tsi><rt-entry><protocol-name>BGP</protocol-name><local-preference>10</local-preference></rt-entry><rt-entry><current-active/><protocol-name>BGP</protocol-name><preference>170</preference><local-preference>220</local-preference><as-path>AS path: 64496 I</as-path><nh><to>192.0.2.2</to></nh><nh><to>192.0.2.3</to><selected-next-hop/></nh><communities><community>64496:1</community></communities></rt-entry></rt></route-table></route-information></rpc-reply>`
 r,e:=parseXML([]byte(s),netip.MustParseAddr("192.0.2.1"))
 if e!=nil||r.Prefix!="192.0.2.0/24"||r.Protocol!="BGP"||r.ASPath!="64496 I"||r.Raw["local_preference"]!="220"||r.Gateway!="192.0.2.3"||r.Raw["communities"]!="64496:1"||r.Raw["active"]!="true"{t.Fatalf("%+v %v",r,e)}
}
func TestXMLStaticIPv6(t *testing.T){
 s:=`<rpc-reply><route-information><route-table><rt><rt-destination>2001:db8::</rt-destination><rt-prefix-length>32</rt-prefix-length><rt-entry><active-tag>*</active-tag><protocol-name>Static</protocol-name></rt-entry></rt></route-table></route-information></rpc-reply>`
 r,e:=parseXML([]byte(s),netip.MustParseAddr("2001:db8::1"));if e!=nil||r.Prefix!="2001:db8::/32"||r.Protocol!="Static"{t.Fatalf("%+v %v",r,e)}
}
func TestXMLNoRoute(t *testing.T){r,e:=parseXML([]byte("<rpc-reply><route-information/></rpc-reply>"),netip.MustParseAddr("192.0.2.1"));if e!=nil||r.Prefix!=""{t.Fatalf("%+v %v",r,e)}}
func TestXMLRejectErrors(t *testing.T){for _,s:=range []string{"","error: permission denied","<rpc-reply><rpc-error/></rpc-reply>","<rpc-reply><route-information><error/></route-information></rpc-reply>","<rpc-reply>"}{if _,e:=parseXML([]byte(strings.TrimSpace(s)),netip.MustParseAddr("192.0.2.1"));e==nil{t.Fatalf("accepted %q",s)}}}
