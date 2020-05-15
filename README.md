# go-dynamic-netcup-ipv6-dns-updater

> Disclaimer: This is my very **first** Golang project. So dont be too harsh.

## All about this project
- Update multiple AAAA records for netcup DNS provider in a programmatic way
- Configuration picked up over env variables
- Can be used together with `cron` for recurring updates

## Target audience

Everyone who wants to expose a pc/ raspberry from a home network and runs into double-NATing problems using IPv4. (and also owns domain from netcup.)

## Build instructions

Build the application yourself:

```bash
go build
```  

## Installation instructions
 
 Provide the configurations over environment variables:
 
```bash
# your netcup domain
export DOMAIN=example.org 

# information from the netcup customer control panel
export CUSTOMERNR=12345  
export APIKEY=abcdefghijklmnooqrstuvwxyz 
export APIPASSWORD=abcdefghijklmnooqrstuvwxyz
 
# comma separated list of hosts (@ = root, * = wildcard)
export HOSTS=*,@ 

./go-dynamic-netcup-ipv6-dns-updater
```

If you want execute this script more than once, consider throwing this command into a shell script and configuring a `cron`
job for recurring execution. Luckily I already prepared everything for you:

````bash
# this commands were tested on raspbian. when you use another linux distribution, the commands may vary

git clone https://github.com/hikkoiri/go-dynamic-netcup-ipv6-dns-updater.git
cd go-dynamic-netcup-ipv6-dns-updater

# build the application
# install go with 'sudo apt install golang-go'
go build

# update the configuation in update.sh
nano update.sh
chmod +x update.sh

# create crontab file (this will run each 5 minutes)
echo "*/5 * * * * $(pwd)/update.sh >> $(pwd)/cron.log 2>&1" >> $(pwd)/crontab
# delete logs each month
echo "0 0 1 * * rm $(pwd)/cron.log" >> $(pwd)/crontab


# configure cron
crontab crontab

# verify cron setup
crontab -l 
tail -f  cron.log
````

Uninstall:

```bash
crontab -r
cd ..
rm -rf go-dynamic-netcup-ipv6-dns-updater
```
