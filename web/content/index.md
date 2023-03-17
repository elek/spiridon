This is a community satellite for the federated Storj network.

## What is this?

This is a <b>lightweight</b> community Satellite for [Storj](https://storj.io) network.

It doesn't store permanent data, but helps to monitor your nodes + share the collected statistics with the community.

For example: a connected Telegram bot can ping you, in case of any Storagenode problems.

## What is a satellite?

Satellite is the metadata server of the federated [Storj](https://storj.io) network. Usually it stores the location the data, while the
connected Storagenodes are hosting the data.

This Satellite is special as it doesn't really store data.

## How can I join to this Satellite

Set the following environment variables for yor Storagenode:

```bash
STORJ_STORAGE2_TRUST_SOURCES="https://www.storj.io/dcs-satellites,storj://1NusSk8HjWppghiWvofEBDqryDaxiALPah8EyRxXWdNkwXg7ai@spiridon.anzix.net:7777"
STORJ_SERVER_USE_PEER_CA_WHITELIST=false
STORJ_HEALTHCHECK_DETAILS=true
```

Note: `STORJ_HEALTHCHECK_DETAILS` is not required, but it helps to maintain a list of [satellites](/satellites) with
reputation (usage) data. This flag will share the connected satellites + scores with the public (including this satellite).

*Warning*: Please carefully check which data is published on this site (address, version, uptime,...). Join only if you are comfortable with transparent data sharing. (Operator email address will never be shared, as we don't store them.)

## How can I get notifications about my Storagenode

Search for `@satellite_spiridon_bot` on Telegram and use `/subscribe <nodeid>`.

Supported commands:

* `/subscribe <nodeid>`: subscribe to any status change related to the given `<nodeid>`
* `/unsubscribe <nodeid>`: unsubscribe for any new status change related to the give node
* `/subscriptions`: list active subscriptions

## How does it differ from original Storj satellites

* It doesn't disqualify you
* It doesn't store real data (only a minimal amount to test actual availability / performance)
* It gives you fast notification (Using Telegram bot)
* It provides freely available data from the connected nodes to everybody
  * nodeID, connectivity information (IP/domain), availability (latency / upload speed / health), free space

Planned resource boundaries:

* Data stored permanently: `<100MB`,
  * typically store only one 4MB piece to monitor retrievability (not yet implemented)
  * Occasional test with bigger files (not yet implemented)
* Planned temporary store (deleted in one day): `<1GB`
  * Only for occasional tests. Planned to provide it per-request of the owner (not yet implemented)
* Planned bandwidth usage `<1GB` / day.
  * an hourly 4Mb upload/download check is `24 * 4 * 2 = 200Mb`) (not yet implemented)
  * Other occasional tests may exceed this (not yet implemented)

## Do you have feature plans

A lot. For example:

* notification bot for other protocols (like ntfy, but it requires code change in satellite protocol)
* better statistics (version distribution, ...)
* measuring and share latency data (maybe from multiple locations) and share the feedback (how good is your Storagenode
  compared to others.)
* single piece audit: store one single piece on everynode (4MB) and constantly check retrievability
* advanced features for owners (metrics from the storagenodes?)
* personal dashboard for node operators

## How does it work technically

This is a full reimplementation of Storj Satellite protocol without using the original code of `storj/storj`. It's quite
easy as we don't need to support any real upload/download/accounting only a very few endpooint to manage nodes.

## Is it a paid satellite?

Nope. But it doesn't really use your resources anyway.

If you would like to get money, join to the Storj Labs
satellite: [https://www.storj.io/dcs-satellites](https://www.storj.io/dcs-satellites) (Enabled, by default)

## Who is the operator of this Satellite?

[Márton Elek](https://github.com/elek). Long-term open source maintained, Apache committer / PMC member, and -- as of
now -- employee of Storj Labs. Running this Satellite is a private effort and independent on Storj Labs. (but expertise
learned during the daily job certainly helped).

# Bug report / feature request / security report

https://github.com/elek/spiridon

