---
RFC: 008
Author: George Georgiev <gngeorgiev.it@gmail.com>
Status: Draft
---

# First class citizen websockets support

## Motivation

There are cases when one's microservices need to be connected with websockets. 
Go kit is the perfect candidate due to its modular nature and variety of tools for microservices.

## Scope

The websockets support should be as easy to use as the JSON over HTTP api. While there are fundamental differences 
between the two, the implementation should focus on minimizing this gap so go-kit users can have the same or 
a very close api.

The websockets api should very well integrate (where possible) with the other parts of go-kit, such as *logging*.


## Implementation

To be defined.


