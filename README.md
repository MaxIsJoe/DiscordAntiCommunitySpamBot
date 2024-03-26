# Community Spam Bot Protector
A bot that prevents scammers from spamming community servers, written in Go.

Created after I got annoyed with scammers joining the [Unitystation](https://www.unitystation.org/) discord server.

This bot only bans new users who just joined the server within an hour, while older users will only get muted.

## How to setup:

### First Steps:
1. Go into config.ini and add a token to the bot that you want to host using this.
2. Setup things like the spam limit from the config.ini file.

### Build Directly using Go:
3. Build the project using `go build`
4. Run the built Application, then let it do its job.

### Build and deploy using Docker:
3. Run `docker build -t antispambot .` to build
4. Run `docker run -p 8080:8080 antispambot` to deploy.

## Support me

https://maxisjoe.xyz/maxfund
