# probe
probe is a small tool that will measure your server response times 

# dependencies
for calculation of various percentiles a stats pkg was used, in order to compile the binary go get it.
```sh
go get github.com/montanaflynn/stats
```
note:
the pkg is not vendored at the moment - hopefully the authors of that pkg will not make any changes that could break the API.

# usage
TODO: add description

# testing
by default tests are run with --race flag.
```sh
make test 
```

for more verbous output set -v flag.
```sh
make test T_FLAGS="-v"
```
