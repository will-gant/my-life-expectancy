# My life expectancy

Go script to benchmark the longevity of the direct descendents of the first individual in a [GEDCOM](https://www.gedcom.org/) family tree file against modal and median death ages for the United Kingdom for each ancestor's year of death. For all ancestors who died in years for which ONS statistics are available (i.e. 1841-2010), the diff from the modal and median death ages is calculated. An average is then calculated that weights each ancestor diff base on their proximity to the first individual. i.e. the diffs for the ancestor's parents have twice the weight of those of their grandparents, which in turn have twice the weight of those of their great-grandparents, etc.

GEDCOM files (with a `.ged` extension) can be exported from a number of genealogy websites, such as ancestry.com. I gradually built my own tree over several months, and in my case I have 68 director ancestors (34 men and women) who died in a year covered by the ONS statistics, roughly covering five generations (i.e. back to great-great-great-grandparents).

I previously implemented a simpler version of this in ruby that used a CSV generated from my GEDCOM file by [Gramps](https://gramps-project.org/), but this resulted in the information on generational proximity being lost. As the ruby GEDCOM libraries that I tried seemed like they themselves had passed on to a better place and seemed unable to read my `.ged` file without exploding, I looked at other languages and found a [nice Go package](https://github.com/iand/gedcom) that seemed to be actively maintained and did the job.

## Data sources

Historical data on UK mortality come in `.csv` files bundled with a [statistical release](https://web.archive.org/web/20221124074230/https://www.ons.gov.uk/peoplepopulationandcommunity/birthsdeathsandmarriages/lifeexpectancies/articles/mortalityinenglandandwales/pastandprojectedtrendsinaveragelifespan) of the UK's Office of National Statistics entitled _Mortality in England and Wales: past and projected trends in average lifespan_, published on 5 July 2022.

I had to manually tweak a couple of tiny details (e.g. header names) that weren't exactly consistent between the male and female files published by the ONS.

## Assumptions/limitations

* Assumes a death date of 1 January where the dataset gives only a year
* Similarly, where only a month and a year are available, assumes the death occured on the 1st of that month
* Where a death date is recorded as a range of years (e.g. `1905-1907`) assumes that the death date is the midway point between the two
* Ignores leap years (i.e. assumes years are all 365 days long)
* Obviously, uses UK death statistics for all ancestors. Apart from the amount of effort that'd be required in obtaining equivalent stats for other countries (assuming they even exist), trying to decide _which_ country's statistics to apply to a given ancestor would be a nightmare. I guess in an ideal world you'd use whichever country they spent the most time in, but suffice to say this is rarely available. Even when locations are given for deaths, births, etc, these may omit the country entirely (e.g. only give a town/city) or use a range of different names (e.g. "England", "United Kingdom" and "UK").

## Usage

Here's the cheerful result that I get using my own family tree:

```console
$ % go run predict-death.go --tree-file tree.ged
===========================================================================================
Stat                                Male               Female             Overall
Difference from Median Death Age    7 years 126 days   5 years 154 days   6 years 135 days
Difference from Modal Age at Death  -4 years 148 days  -4 years 226 days  -4 years 188 days
===========================================================================================
```

## Tests

```console
$ go test
PASS
ok  	predict-death	0.153s
```
