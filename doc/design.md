# Additional Design Notes

This document contains some additional design notes and details.


## Overview

The primary components of controller bus are:

 - **Config**: an object that configures a controller at construct time.
 - **Controller**: state machine / goroutine processing Directives on a bus.
 - **Bus**: a channel to connect together multiple Controllers.
 - **Factory**: contains controller implementation metadata and constructors.
 - **Directive**: an ongoing request for data or desired state.
 - **Resolver**: concurrent process(es) computing values to satisfy a directive.

Controllers are started attached to a common Bus. They can be directly attached
or loaded with directives to the "loading controller." A directive to load and
start a controller might be resolved by fetching code from the network and
loading a dynamic library, for example. Controllers have a single entrypoint
Goroutine but can spawn other routines as needed.

Directive objects can be attached to a Bus, where they are passed to all running
controllers for handling. Directives are de-duplicated, and reference counting
is used to determine when a directive can be canceled and released.

The controllerbus system manages starting and stopping resolvers yielded by the
controller handlers. A resolver executes until the directive has the desired
number of values, or the directive is canceled. Resolvers can be started and
stopped multiple times in the life-span of a directive.

A "Value" is an opaque object attached to a Directive, which will ultimately be
returned to the originator of the Directive. Bounded directives accept a limited
number of values before canceling remaining resolvers. Values can be expired,
and if the desired value count drops below a threshold, the resolvers will be
restarted until new values are found. A bounded directive with a value limit of
1 is sometimes referred to as a "singleton" in this document and the codebase.

