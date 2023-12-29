
* Figure out integration wih wg-quick and wireguard.exe. Do we write state to /etc/wiregaurd/wg.conf and C:\Program Files\WireGuard\Data\Configurations\wg0.conf.dpapi and let systemd and services call them at start up?
    * On wg-quick has a save on shutdown so the server can just deal save wg config and have it saved by wg quick service (should print a reminder if nothig under /etc/wireguard?). Windows probably does the same 
    * For client creating the interfaces is different in each os so for now just emit a wireguard config and hav them enter that. 
* Ip tables integration. Do we check/ensure forrard is not blocked on server/hub
* Encrypion of incoming requests. Man in the middle attach otherwise? Put public key in dns record. Sign with ens account to much?
* Deletion of OTOP tokens? 
* Friendly names?
* Is pregenerating peers and keeping them on phone simpler.