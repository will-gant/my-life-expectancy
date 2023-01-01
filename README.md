# My life expectancy

Ruby script to benchmark the longevity of 68 direct ancestors of mine who died between 1843 and 1998, roughly covering five generations (back to great-great-great grandparents).

## Data sources

The data on my direct ancestors comes from my family tree on ancestry.com, and was veeeery gradually collated through a mixture of automated processes and manual effort! I then exported this as a `.ged` file and used the desktop app from the open source family tree software [Gramps](https://gramps-project.org/) to generate a `.csv` file that included only direct ancestors.

Historical data on UK mortality come `.csv` files bundled with a [statistical release](https://web.archive.org/web/20221124074230/https://www.ons.gov.uk/peoplepopulationandcommunity/birthsdeathsandmarriages/lifeexpectancies/articles/mortalityinenglandandwales/pastandprojectedtrendsinaveragelifespan) of the UK's Office of National Statistics entitled _Mortality in England and Wales: past and projected trends in average lifespan_, published on 5 July 2022.

## Limitations

* Assumes a death date of 1 January where the dataset gives only a year
* Ignores leap years (i.e. treats all years as if they are 31,536,000 seconds long)
* Uses UK death statistics even though some ancestors lived part or all of their lives in other countries (e.g. South Africa)
* To be consistent with the UK death statistics we're comparing with, filters for ancestors who died aged ten or older. Reasonably confident no direct ancestor of mine managed to reproduce and then die before their tenth birthday in any case!

## Usage

```console
$ ruby lib/predict_death.rb --ancestors direct-ancestors.csv --male-death-stats male_death_stats.csv --female-death-stats female_death_stats.csv

male ancestors in the provided dataset lived 3.57 fewer years than the UK's male modal age of death in the year they died (18/34 outlived the mode)
male ancestors in the provided dataset lived 13.72 more years than the UK's male median age of death in the year they died (29/34 outlived the median)
Calculated from 34 male ancestors who died between 1849 and 1993

female ancestors in the provided dataset lived 3.29 fewer years than the UK's female modal age of death in the year they died (20/34 outlived the mode)
female ancestors in the provided dataset lived 10.36 more years than the UK's female median age of death in the year they died (27/34 outlived the median)
Calculated from 34 female ancestors who died between 1843 and 1998
```

## Tests

```console
$ bundle install
$ bundle exec rspec

.............

Finished in 0.00802 seconds (files took 0.13782 seconds to load)
13 examples, 0 failures
```
