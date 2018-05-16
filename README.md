# Funky

The best functional language ever.

## Short usage manual

Funky is a purely functional language, with no built-in notion of side-effects. Those are handled
by special programs called interpreters. You can try out the `funkycmd` interpreter located in the
[interpreters/funkycmd](interpreters/funkycmd) folder.

To do that, you need to:

1. Install [Go](https://golang.org/) and [set up the $PATH](https://golang.org/doc/install).
2. Download this repo: `go get github.com/faiface/funky`.
3. Navigate to the repo folder.
4. `cd interpreters/funkycmd`
5. `go install`
6. `cd ../../test`
7. `funkycmd ../stdlib/*.fn ../stdlib/funkycmd/*.fn test.fn`
8. There you go!
