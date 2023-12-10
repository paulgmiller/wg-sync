[![Go](https://github.com/paulgmiller/wg-sync/actions/workflows/go.yml/badge.svg)](https://github.com/paulgmiller/wg-sync/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/paulgmiller/wg-sync)](https://goreportcard.com/report/github.com/paulgmiller/wg-sync)


# What is this?

A way to set up a [wireguard](https://www.wireguard.com/) network. Having a simple network you can depend on to be secure is pretty great. Open up that un-auth website. Share files.
SSH? Maybe just regular remote shell. Its all in the network! Like when you friends ran ethernet to eachothers rooms in college.

You should probably just use [tailscale](https://tailscale.com/) they've made a really great product. 
I used it. It's great. Works without much fuss. You Should probably use that.

## But...
While tailscale's e2e encyption is great they do control your peers. So at anypoint they could just shove another peer into your vpn and party down.
That's probably fine. You've gotta trust somebody [right](https://www.bing.com/ck/a?!&&p=7e3d9888db2fefcaJmltdHM9MTcwMTk5MzYwMCZpZ3VpZD0wNTdhN2Y4Yi01MTc4LTZiOGQtMTk3ZC02YzVlNTBjMjZhOWMmaW5zaWQ9NTE4NA&ptn=3&ver=2&hsh=3&fclid=057a7f8b-5178-6b8d-197d-6c5e50c26a9c&psq=trusting+trust+acm&u=a1aHR0cHM6Ly93d3cuY3MuY211LmVkdS9-cmRyaWxleS80ODcvcGFwZXJzL1Rob21wc29uXzE5ODRfUmVmbGVjdGlvbnNvblRydXN0aW5nVHJ1c3QucGRm&ntb=1)?
Maybe you're crazy paranoid or work for a company that decides they can't deal with that adependenc. 

## What then.
Setting up manual wireguard network is [totally doable](https://www.wireguard.com/quickstart/) but it is not convenient.
What if you've got a new machine and can't get to an existing one. Even if got one do you really want copy keys across through whatsapp or sms like a chump?

Instead what if you could download a tool on the new machine point it at a public dns/ip and use a two factor from your phone?

## What do I need?
A server (virtual or otherwise) with a public ip. Need to be able to open udp ports on the public ip.

## Then what happens? 
On your public server setup wg-sync serve. Get an [TOTP token](https://github.com/sec51/twofactor) from that sever either locally or from other machine on wireguard network.  On a new machine run wg-sync add and give it the the public ip/dns and token. Bam connected.

Under the covers we send udp packet with the udp token and added machines public ip. Once recived that is done get a new token. 

## Is this secure?
Maybe not I probably fucked something up. The TOTP token should be encypted or you could be man in the middled. We could encyrpt with the public key of the public server.
But thats harder to carry around/type in. Could put it in txt of DNS. TODO I guess. 

## Surely somone has done this
Maybe? Need to read more about [dsnet](https://github.com/naggie/dsnet) and [subspace](https://github.com/subspacecommunity/subspace). Lots of other [neat stuff here](https://github.com/cedrickchee/awesome-wireguard)

## Anything else?
Yes i've been drinking bourbon. Why do you ask? 
