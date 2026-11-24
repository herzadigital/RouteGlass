package juniper

import (
	"context"
	"errors"
	"fmt"
	"github.com/herzadigital/routeglass/internal/config"
	"github.com/herzadigital/routeglass/internal/mikrotik"
	"net/netip"
	"os/exec"
	"encoding/xml"
 "io"
	"strings"
)

type Client struct{ cfg config.JuniperConfig }

var ErrDisabled = errors.New("BGP lookup is not enabled")

func New(c config.JuniperConfig) *Client { return &Client{c} }

func (c *Client) Lookup(ctx context.Context, target string) (mikrotik.Result, error) {
	ip, err := netip.ParseAddr(target)
	if err != nil {
		return mikrotik.Result{}, errors.New("target must be a valid IPv4 or IPv6 address")
	}
	if c.cfg.Address == "" || c.cfg.Username == "" || c.cfg.PrivateKeyFile == "" {
		return mikrotik.Result{}, errors.New("invalid Juniper configuration")
	}
	cmd := exec.CommandContext(ctx, "ssh", "-i", c.cfg.PrivateKeyFile, "-p", fmt.Sprint(c.cfg.Port), "-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+c.cfg.KnownHostsFile, "-o", "ConnectTimeout=5", c.cfg.Username+"@"+c.cfg.Address, "show route "+ip.String()+" extensive | display xml | no-more")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return mikrotik.Result{}, fmt.Errorf("Juniper lookup failed: %s", strings.TrimSpace(string(out)))
	}
	return parseXML(out, ip)
}

type routeEntry struct {
 Active *struct{} `xml:"current-active"`
 Tag string `xml:"active-tag"`
 Protocol string `xml:"protocol-name"`
 Preference string `xml:"preference"`
 ASPath string `xml:"as-path"`
 LocalPref string `xml:"local-preference"`
 PeerAS string `xml:"peer-as"`
 Communities struct { Values []struct { Value string `xml:",chardata"` } `xml:",any"` } `xml:"communities"`
 NH []struct { To string `xml:"to"`; Selected *struct{} `xml:"selected-next-hop"` } `xml:"nh"`
}
func parseXML(out []byte, ip netip.Addr)(mikrotik.Result,error){
 res:=mikrotik.Result{Target:ip.String()}
 var reply struct {
 XMLName xml.Name `xml:"rpc-reply"`
 Info *struct { Tables []struct { Routes []struct {
 Destination string `xml:"rt-destination"`
 Length string `xml:"rt-prefix-length"`
 Entries []routeEntry `xml:"rt-entry"`
 } `xml:"rt"` } `xml:"route-table"` } `xml:"route-information"`
 }
 if err:=xml.Unmarshal(out,&reply);err!=nil{return res,fmt.Errorf("invalid Junos XML: %w",err)}
 dec:=xml.NewDecoder(strings.NewReader(string(out)))
 for {tok,err:=dec.Token();if err==io.EOF{break};if err!=nil{return res,err};if el,ok:=tok.(xml.StartElement);ok&&(el.Name.Local=="rpc-error"||el.Name.Local=="error"){return res,errors.New("Junos route query failed")}}
 if reply.Info==nil{return res,errors.New("missing Junos route information")}
 best:=-1
 for _,table:=range reply.Info.Tables{for _,rt:=range table.Routes{
 p:=rt.Destination;if !strings.Contains(p,"/"){p+="/"+rt.Length};prefix,err:=netip.ParsePrefix(p);if err!=nil{return mikrotik.Result{},err};if !prefix.Contains(ip)||prefix.Bits()<=best{continue}
 if len(rt.Entries)==0{return mikrotik.Result{},errors.New("missing route entry")}
 e:=rt.Entries[0];for _,v:=range rt.Entries{if v.Active!=nil||v.Tag=="*"{e=v;break}}
 if e.Protocol==""{return mikrotik.Result{},errors.New("missing route protocol")}
 best=prefix.Bits();res=mikrotik.Result{Target:ip.String(),Prefix:prefix.Masked().String(),Protocol:e.Protocol,Distance:e.Preference,ASPath:strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(strings.SplitN(e.ASPath,"\n",2)[0]),"AS path:")),Raw:map[string]string{"local_preference":e.LocalPref,"peer_as":e.PeerAS,"active":fmt.Sprint(e.Active!=nil||e.Tag=="*")}}
 for _,nh:=range e.NH{if nh.Selected!=nil{res.Gateway=nh.To;break}};if res.Gateway==""&&len(e.NH)==1{res.Gateway=e.NH[0].To}
 var communities []string;for _,v:=range e.Communities.Values{if s:=strings.TrimSpace(v.Value);s!=""{communities=append(communities,s)}};res.Raw["communities"]=strings.Join(communities," ")
 }}
 return res,nil
}
