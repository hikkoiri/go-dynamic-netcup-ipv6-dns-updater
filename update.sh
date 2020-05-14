# your netcup domain
export DOMAIN=example.org

# information from the netcup customer control panel
export CUSTOMERNR=12345
export APIKEY=abcdefghijklmnooqrstuvwxyz
export APIPASSWORD=abcdefghijklmnooqrstuvwxyz

# comma separated list of hosts (@ = root, * = wildcard)
export HOSTS=*,@

./go-dynamic-netcup-ipv6-dns-updater