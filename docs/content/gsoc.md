---
title: Google Summer of Code
weight: 7
---

# Google Summer of Code

## Introduction

Google Summer of Code (GSoC) is a global program focused on bringing more student developers into open-source software development.
Students work with an open source organization on a 12-week programming project during their break from school.

Take a look at the [official GSoC website](https://summerofcode.withgoogle.com/) for more information.

## Forwarder's Participation

Forwarder is participating in GSoC 2024.
We are looking for students who are passionate about open source and want to contribute to the project.

## Ideas

The following list of ideas is not exhaustive, and we are open to new ideas as well.
If you have a project idea feel free to reach out to us.
The projects are sorted by difficulty, with the easiest ones at the top.

### UI (or TUI) for request analysis in Forwarder proxy

The goal of this project is to create a user interface for viewing and analyzing the recorded requests.
We are open to both web-based and terminal-based UIs.
There are terminal-based UI libraries like [bubbletea](https://github.com/charmbracelet/bubbletea) that allow for creating very exciting TUIs in Go.
To do that we'd need to record the requests, and provide a way view them.

#### Expected outcomes

* In memory request recording
* UI for viewing the recorded requests

#### Skills required/preferred

* Go
* TUI or web development
* SQL

#### Possible mentors

* [Michał Matczuk](github.com/mmatczuk)
* [Hubert Grochowski](github.com/Choraden)

#### Expected size of project

175 hours, easy

### Request recording and analysis in Forwarder proxy

The goal of this project is to add support for recording and analyzing requests in the Forwarder proxy.
To do that we'd need to record the requests in a columnar format like Apache Arrow, and then provide a way to analyze the data.
We could use a SQL LLM to convert user queries into SQL queries and then use a columnar database to analyze the data.

#### Expected outcomes

* Request recording in a columnar format
* SQL interface for analyzing the recorded requests
* Text-based UI for analyzing the recorded requests

#### Skills required/preferred

* Go
* SQL
* Machine learning

#### Possible mentors

* [Michał Matczuk](github.com/mmatczuk)
* [Hubert Grochowski](github.com/Choraden)

#### Expected size of project

175 hours, intermediate

### gRPC support in Forwarder proxy

The goal of this project is extending the Forwarder proxy with dedicated support for gRPC. 
We want to go beyond the HTTP/2 support and provide an idiomatic way for RPC traffic analysis and modification.
To do that we'd need to review the existing tools and libraries and see how we can integrate them into Forwarder. 

#### Expected outcomes

* Support for encrypted and unencrypted gRPC traffic
* JavaScript based modification of RPC traffic
* RPC call based logging
* (Optional) recording and replaying RPC calls 
* (Optional) integration with UI

#### Skills required/preferred

* Go
* gRPC
* HTTP/2

#### Possible mentors

* [Michał Matczuk](github.com/mmatczuk)
* [Hubert Grochowski](github.com/Choraden)

#### Expected size of project

175 hours, intermediate

### Wireguard integration in Forwarder proxy

The goal of this project is to incorporate a WireGuard Server into the Forwarder proxy.
This integration aims to significantly streamline and enhance mobile device integration with Forwarder.
MITMproxy added a similar feature in October 2022, see [here](https://mitmproxy.org/posts/wireguard-mode/) for more information.

Our implementation, can be more integrated and robust. 
Utilizing [Go WireGuard](https://github.com/WireGuard/wireguard-go), extensively used by entities like TailScale, we will seamlessly integrate it directly into the Forwarder proxy. This approach not only ensures better performance but also optimizes the overall usage of the proxy.

#### Expected outcomes

* WireGuard Server integration
* Seamless mobile device integration

#### Skills required/preferred

* Go
* WireGuard
* Networking
* Mobile development

#### Possible mentors

* [Michał Matczuk](github.com/mmatczuk)
* [Hubert Grochowski](github.com/Choraden)

#### Expected size of project

350 hours, hard

### HTTP/3 MITM support in Forwarder proxy

HTTP/3 is the next version of the HTTP protocol and is designed to improve the performance of web traffic.
This project aims to add support for HTTP/3 in the Forwarder proxy, including MITM support.

#### Expected outcomes

* Support for HTTP/3
* MITM support for HTTP/3
* Support for upstream forwarding of HTTP/3

#### Skills required/preferred

* Go
* HTTP/3
* Networking

#### Possible mentors

* [Michał Matczuk](github.com/mmatczuk)
* [Hubert Grochowski](github.com/Choraden)

#### Expected size of project

350 hours, hard

## How to Apply

To apply for GSoC with Forwarder, you need to follow these steps:

* Familiarize yourself with the project
  * Read the documentation
  * Install and run Forwarder
  * Read the source code
* Try sending a pull request to fix a bug or add a feature
* Contact us
* Write your proposal with help from your mentors!
* Submit your proposal to Google
  All applications must go through [Google's application system](https://summerofcode.withgoogle.com/); we can't accept any application unless it is submitted there.

### What to include in your proposal

#### About me

* Name (and nicknames like your github and irc usernames)
* University / program / year / expected graduation date
* Contact info (email, phone, etc.)
* Time zone
* Link to a resume (if you want)
* Code contribution (link to a pull request)

#### Project information

* Project Abstract
* Detailed description
* Weekly timeline (what you plan to do each week)
* Any other information you think is relevant

#### Why you?

* What makes you the best person to work on this project?
* What relevant experience do you have?

#### Why us?

* Why do you want to work with us?
* What do you hope to get out of the experience?
