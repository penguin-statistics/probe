<img src="https://penguin.upyun.galvincdn.com/logos/penguin_stats_logo.png"
alt="Penguin Statistics - Logo"
width="96px" />

# Penguin Statistics - `probe`
[![Status](https://img.shields.io/badge/status-production-green)](#readme)
[![Language](https://img.shields.io/badge/using-Go-%2300add8?logo=go)](#readme)
[![Go Version](https://img.shields.io/github/go-mod/go-version/penguin-statistics/probe)](https://github.com/penguin-statistics/probe/blob/main/go.mod)
[![GoDoc](https://pkg.go.dev/badge/github.com/penguin-statistics/probe)](https://pkg.go.dev/github.com/penguin-statistics/probe)
[![Go Report Card](https://goreportcard.com/badge/github.com/penguin-statistics/probe)](https://goreportcard.com/report/github.com/penguin-statistics/probe)
[![License](https://img.shields.io/github/license/penguin-statistics/probe)](https://github.com/penguin-statistics/probe/blob/main/LICENSE)
[![Last Commit](https://img.shields.io/github/last-commit/penguin-statistics/probe)](https://github.com/penguin-statistics/probe/commits/main)
[![GitHub Actions Status](https://github.com/penguin-statistics/probe/actions/workflows/build.yml/badge.svg)](https://github.com/penguin-statistics/probe/actions/workflows/build.yml)

This is the **probe** project repository for the [Penguin Statistics](https://penguin-stats.io/?utm_source=github) website.

The `probe` service here is designed to only serve as a simple counter for some of the metrics that our user may generate when interact with the UI of our website, and we will not and will also not able to identify and track our users using this service.

## Overview
The `probe` service is exposed with a `WebSocket` endpoint which is backward compatible with clients which don't support `WebSocket`, as we encode the initial `visit` event into query string when we tries to use `new WebSocket` to establish a new connection with the server. If the http request was made to the server but the client was unable to establish a valid websocket connection for any reason, their visit event would still be counted due to the fact that the technological data has been already transmitted to the server. This way this service will offer delightful compatibility, including those who may have `<noscript>` on and are not able to visit our website.

By providing the initial and very basic information about a "visit" event we would be (finally!) able to get a near-perfect count of the amount of users who uses our website, to give us a better shape of understanding about the user base we had which enables us to provide better optimizations for all of our users.

### Choice of WebSocket
Afterwards, some of the interactions the user performed with UI would be encoded into [`protobuf`](https://developers.google.com/protocol-buffers), a very compact message format which allows us to reduce the network footprint to the scale of a single byte. This is to save bandwidth cost on both the user's side and our side, as well to significantly improve the transport efficiency of messages. The reason of we choosing websocket is that:
1. Transport itself is lightweight after the connection established
2. Very small message transport footprint (comparing to http where even http/2 would need lots of space to represent the request headers)
   - Especially for very tiny data chunks that are being transmitted relatively frequently
3. Allows simple heartbeat detection (automatic `PONG` reply as per [RFC6455 Standard - Section 5.5.2 - Ping](https://www.rfc-editor.org/rfc/rfc6455.html#section-5.5.2))
4. Re-use of initial visit event data without the introduction of a bulky **session** system
5. Suitable for poor network quality (easy retransmission - just push again everything server not received, comparing to http requests which even may rate limit the client when there's too much to be retransmitted)

### User Privacy
The `visit` event currently, consists of three elements: Client Version (e.g. `v3.4.1`), Platform (e.g. `web` or `app:ios`), and a user-side randomly generated user ID that is stored privately on the visitor's device, only serves as a purpose to de-duplicate the possible repeated visits from one single specific device to our website. The randomly generated ID here, is generated on client-side, does not link to any third-party trackers, safely stored _(in `LocalStorage` so it won't be sent automatically and shall only be able to read by codes from Penguin Statistics, in a safety-modal matter)_ and _will_ expire (to be re-generated) after 180 days.

The `probe` service is designed to protect our user's privacy and will not upload any sensitive information to the server.

## Maintainers
This project has mainly being maintained by the following contributors (in alphabetical order):
- [GalvinGao](https://github.com/GalvinGao)

> The full list of active contributors of the *Penguin Statistics* project can be found at the [Team Members page](https://penguin-stats.io/about/members) of the website.

## How to contribute?
Our contribute guideline can be found at [Penguin Developers](https://developer.penguin-stats.io). PRs are always more than welcome!