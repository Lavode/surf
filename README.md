# Succinct Range Filters (SuRF)

[Succinct Range Filters](http://www.cs.cmu.edu/~huanche1/publications/surf_paper.pdf) (SuRF) 
is a data structure providing probabilistic exact- and range- membership checks. 

This implementation is done as part of a MSc lecture on data structures at the University of Fribourg. There's of course nothing stopping you from using it in a project of yours, but buyers beware. :)

## Running tests and benchmarks

To run all tests run, from the root directory:
```bash
go test ./...
```

There are also some benchmarks, which can be run from the root directory:
```
go test ./... -bench=.
```

## Licensing

Unless indicated otherwise, all parts of this project are licensed under the Apache 2.0 license.
