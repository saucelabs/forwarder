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

We have a list of project ideas that we think would be beneficial for the project.
However, we are open to new ideas as well.
If you have a project idea that you think would be beneficial for the project, feel free to reach out to us.

### Project Ideas

#### Request recording

Request recording, preferably to a columnar format like Apache Arrow, opens up a lot of possibilities for analysis and debugging.

##### AI-based request analysis

The goal of this project is to use AI to analyze the recorded requests and provide insights into the traffic.
We can use a SQL LLM to convert user queries into SQL queries and then use a columnar database to analyze the data.

Another possibility is traffic prediction, for instance we would be interested in predicting traffic spikes to allocate resources accordingly.

##### UI or TUI for request analysis

The goal of this project is to create a user interface for viewing and analyzing the recorded requests.
There are terminal-based UI libraries like [bubbletea](https://github.com/charmbracelet/bubbletea) that allow for creating exciting TUIs in Go.

#### gRPC visualization and modification

The goal of this project is adding dedicated support for gRPC in Forwarder.
It would allow for visualizing and modifying gRPC traffic.
We can easily add JavaScript based modification support for gRPC traffic.  

#### Wireguard integration in Forwarder proxy

The goal of this project is to incorporate a WireGuard Server into the Forwarder proxy.
This integration aims to significantly streamline and enhance mobile device integration with Forwarder.
MITMproxy added a similar feature in October 2022, see [here](https://mitmproxy.org/posts/wireguard-mode/) for more information.

Our implementation, can be more integrated and robust. 
Utilizing [Go WireGuard](https://github.com/WireGuard/wireguard-go), extensively used by entities like TailScale, we will seamlessly integrate it directly into the Forwarder proxy. This approach not only ensures better performance but also optimizes the overall usage of the proxy.

#### HTTP/3 MITM support in Forwarder proxy

HTTP/3 is the next version of the HTTP protocol and is designed to improve the performance of web traffic.
This project aims to add support for HTTP/3 in the Forwarder proxy, including MITM support.

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
