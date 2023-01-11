# My life expectancy

## Contents
1. [Background](#background)
1. [Data sources](#data-sources)
1. [Assumptions and limitations](#assumptions-and-limitations)
1. [Usage](#usage)
1. [Tests](#tests)

## Background

Go script to benchmark the longevity of the direct descendents of the first individual in a [GEDCOM](https://www.gedcom.org/) family tree file (i.e. me) against modal and median death ages for the United Kingdom for each ancestor's year of death. For all ancestors who died in years for which ONS statistics are available (i.e. 1841-2010), the diff from the modal and median death ages is calculated. An average is then calculated that weights each ancestor's diff based on their proximity to the first individual in the tree. i.e. the diffs for my grandparents have twice the weight of those of my great-grandparents, which in turn have twice the weight of those of my great-great-grandparents, etc. As the number of individuals doubles with each generation you go back, that means that _collectively_ the diffs for a given generation carry the same weight as those for any another.

GEDCOM files (with a `.ged` extension) can be exported from a number of genealogy websites, such as ancestry.com. I gradually built my own tree over several months, and in my case I have 78 director ancestors who died in a year covered by the ONS statistics, stretching back seven generations (i.e. back to great-great-great-great-great-grandparents).

I previously implemented a simpler version of this in ruby that used a CSV generated from my GEDCOM file by [Gramps](https://gramps-project.org/), but this resulted in the information on generational proximity being lost. As the ruby GEDCOM libraries that I tried seemed like they themselves had passed on to a better place and seemed unable to read my `.ged` file without exploding, I looked at other languages and found a [nice Go package](https://github.com/iand/gedcom) that seemed to be actively maintained and did the job.

<a href="#contents">Back to top</a>
## Data sources

Historical data on UK mortality come in `.csv` files bundled with a [statistical release](https://web.archive.org/web/20221124074230/https://www.ons.gov.uk/peoplepopulationandcommunity/birthsdeathsandmarriages/lifeexpectancies/articles/mortalityinenglandandwales/pastandprojectedtrendsinaveragelifespan) of the UK's Office of National Statistics entitled _Mortality in England and Wales: past and projected trends in average lifespan_, published on 5 July 2022.

I had to manually tweak a couple of tiny details (e.g. header names) that weren't exactly consistent between the male and female files published by the ONS.

<a href="#contents">Back to top</a>
## Assumptions and limitations

* Assumes a death date of 1 January where the dataset gives only a year
* Similarly, where only a month and a year are available, assumes the death occured on the 1st of that month
* Excludes ancestors for whom no birth or death year is available (a range of years is acceptable - e.g. `1905-1907`).
* Where a death date is recorded as a range of years, assumes that the death date is the day that falls halfway between the two
* Ignores leap years (i.e. assumes years are all 365 days long)
* Obviously, uses UK death statistics for all ancestors. Apart from the amount of effort that'd be required in obtaining equivalent stats for other countries (assuming they even exist), trying to decide _which_ country's statistics to apply to a given ancestor would be a nightmare. I guess in an ideal world you'd use whichever country they spent the most time in, but suffice to say this is rarely available. Even when locations are given for deaths, births, etc, these may omit the country entirely (e.g. only give a town/city) or use a range of different names (e.g. "England", "United Kingdom" and "UK").

<a href="#contents">Back to top</a>
## Usage

By default the script just outputs the results in a human-readable format. The optional `--csv` flag can be passed with a desired filename in order to generate a .csv file in which all time durations are given as a number of days (which is easier for people to manipulate in Excel or whatever).

```
$ go run predict-death.go --tree-file tree.ged [--csv somefilename.csv]
```

Here's the cheerful result that I get using my own family tree:

```console
$ go run predict-death.go --tree-file tree.ged
===========================================================================================
Longevity statistics for the direct ancestors of William Norman Gant
===========================================================================================
Stat                                Male               Female             Overall
Difference from Median Death Age    7 years 126 days   5 years 154 days   6 years 135 days
Difference from Modal Age at Death  -4 years 148 days  -4 years 226 days  -4 years 188 days
===========================================================================================
Year  Generations removed from subject  Gender  Age at death       Median Death Age Diff  Modal Death Age Diff  Modal Death Age    Median Death Age
1998  2                                 f       88 years 106 days  +5 years 234 days      +2 years 4 days       86 years 102 days  82 years 237 days
1993  2                                 m       78 years 16 days   +1 years 279 days      -1 years 89 days      79 years 105 days  76 years 102 days
1986  2                                 m       77 years 139 days  +2 years 242 days      +0 years 183 days     76 years 321 days  74 years 262 days
1983  3                                 f       94 years 27 days   +13 years 210 days     +9 years 144 days     84 years 248 days  80 years 182 days
1980  2                                 f       63 years 268 days  -16 years 104 days     -20 years 93 days     83 years 361 days  80 years 7 days
1966  3                                 m       81 years 6 days    +9 years 101 days      +5 years 346 days     75 years 25 days   71 years 270 days
1962  3                                 f       91 years 140 days  +13 years 213 days     +9 years 239 days     81 years 266 days  77 years 292 days
1952  3                                 f       79 years 325 days  +3 years 107 days      +0 years 63 days      79 years 262 days  76 years 218 days
1947  4                                 m       90 years 91 days   +20 years 84 days      +14 years 95 days     75 years 361 days  70 years 7 days
1946  3                                 m       59 years 60 days   -11 years 27 days      -17 years 221 days    76 years 281 days  70 years 87 days
1939  4                                 f       77 years 357 days  +5 years 87 days       +0 years 62 days      78 years 54 days   72 years 270 days
1935  4                                 m       78 years 239 days  +10 years 247 days     +3 years 17 days      75 years 222 days  67 years 357 days
1935  3                                 f       50 years 62 days   -21 years 259 days     -27 years 87 days     77 years 149 days  71 years 321 days
1932  3                                 m       61 years 183 days  -5 years 262 days      -13 years 335 days    75 years 153 days  67 years 80 days
1929  4                                 m       86 years 25 days   +21 years 117 days     +10 years 219 days    75 years 171 days  64 years 273 days
1926  4                                 f       82 years 129 days  +11 years 305 days     +5 years 283 days     76 years 211 days  70 years 189 days
1925  4                                 m       65 years 28 days   +0 years 223 days      -10 years 242 days    75 years 270 days  65 years 251 days
1923  3                                 m       54 years 294 days  -11 years 103 days     -20 years 355 days    75 years 284 days  66 years 32 days
1923  5                                 f       85 years 217 days  +15 years 199 days     +8 years 174 days     77 years 43 days   70 years 18 days
1922  4                                 f       86 years 13 days   +17 years 160 days     +8 years 273 days     77 years 105 days  68 years 218 days
1919  6                                 f       85 years 185 days  +19 years 61 days      +8 years 295 days     76 years 255 days  66 years 124 days
1918  5                                 f       89 years 335 days  +30 years 244 days     +13 years 306 days    76 years 29 days   59 years 91 days
1918  4                                 f       76 years 337 days  +17 years 246 days     +0 years 308 days     76 years 29 days   59 years 91 days
1911  5                                 m       84 years 207 days  +24 years 328 days     +10 years 255 days    73 years 317 days  59 years 244 days
1911  5                                 f       84 years 226 days  +20 years 267 days     +9 years 289 days     74 years 302 days  63 years 324 days
1910  4                                 m       73 years 308 days  +13 years 108 days     +1 years 268 days     72 years 40 days   60 years 200 days
1910  5                                 f       82 years 353 days  +17 years 343 days     +8 years 36 days      74 years 317 days  65 years 10 days
1909  5                                 f       89 years 132 days  +26 years 27 days      +14 years 216 days    74 years 281 days  63 years 105 days
1904  5                                 f       79 years 317 days  +19 years 131 days     +5 years 328 days     73 years 354 days  60 years 186 days
1904  5                                 m       76 years 345 days  +21 years 13 days      +5 years 218 days     71 years 127 days  55 years 332 days
1902  4                                 f       52 years 275 days  -7 years 181 days      -21 years 104 days    74 years 14 days   60 years 91 days
1901  4                                 f       67 years 74 days   +8 years 111 days      -6 years 207 days     73 years 281 days  58 years 328 days
1900  5                                 m       80 years 78 days   +28 years 31 days      +9 years 276 days     70 years 167 days  52 years 47 days
1899  5                                 f       74 years 150 days  +17 years 165 days     +1 years 362 days     72 years 153 days  56 years 350 days
1899  4                                 f       37 years 282 days  -19 years 68 days      -34 years 236 days    72 years 153 days  56 years 350 days
1899  5                                 m       65 years 184 days  +13 years 214 days     -5 years 38 days      70 years 222 days  51 years 335 days
1899  5                                 f       78 years 138 days  +21 years 153 days     +5 years 350 days     72 years 153 days  56 years 350 days
1898  5                                 f       86 years 26 days   +27 years 326 days     +13 years 209 days    72 years 182 days  58 years 65 days
1896  4                                 m       54 years 170 days  +0 years 87 days       -17 years 45 days     71 years 215 days  54 years 83 days
1895  5                                 m       74 years 253 days  +22 years 27 days      +3 years 337 days     70 years 281 days  52 years 226 days
1894  4                                 m       47 years 363 days  -7 years 151 days      -24 years 16 days     72 years 14 days   55 years 149 days
1894  5                                 m       66 years 209 days  +11 years 60 days      -5 years 170 days     72 years 14 days   55 years 149 days
1892  4                                 f       31 years 63 days   -25 years 225 days     -39 years 83 days     70 years 146 days  56 years 288 days
1891  5                                 f       63 years 349 days  +8 years 298 days      -5 years 202 days     69 years 186 days  55 years 51 days
1888  5                                 f       80 years 231 days  +23 years 162 days     +8 years 56 days      72 years 175 days  57 years 69 days
1888  4                                 m       71 years 332 days  +19 years 84 days      +2 years 289 days     69 years 43 days   52 years 248 days
1888  5                                 f       84 years 62 days   +26 years 358 days     +11 years 252 days    72 years 175 days  57 years 69 days
1885  6                                 m       80 years 287 days  +30 years 145 days     +11 years 222 days    69 years 65 days   50 years 142 days
1885  5                                 m       80 years 244 days  +30 years 102 days     +11 years 179 days    69 years 65 days   50 years 142 days
1883  5                                 f       76 years 300 days  +22 years 311 days     +3 years 103 days     73 years 197 days  53 years 354 days
1881  5                                 m       70 years 71 days   +19 years 261 days     +0 years 228 days     69 years 208 days  50 years 175 days
1879  6                                 f       82 years 346 days  +29 years 259 days     +11 years 91 days     71 years 255 days  53 years 87 days
1879  6                                 m       80 years 76 days   +31 years 259 days     +9 years 270 days     70 years 171 days  48 years 182 days
1878  5                                 m       73 years 120 days  +27 years 117 days     +1 years 314 days     71 years 171 days  46 years 3 days
1877  6                                 m       85 years 285 days  +37 years 348 days     +14 years 26 days     71 years 259 days  47 years 302 days
1877  5                                 m       68 years 251 days  +20 years 314 days     -3 years 8 days       71 years 259 days  47 years 302 days
1876  6                                 f       69 years 356 days  +18 years 181 days     -2 years 227 days     72 years 218 days  51 years 175 days
1874  6                                 f       60 years 112 days  +9 years 313 days      -12 years 205 days    72 years 317 days  50 years 164 days
1874  6                                 f       83 years 353 days  +33 years 189 days     +11 years 36 days     72 years 317 days  50 years 164 days
1872  5                                 m       49 years 268 days  +3 years 188 days      -22 years 359 days    72 years 262 days  46 years 80 days
1869  5                                 m       58 years 35 days   +13 years 130 days     -13 years 267 days    71 years 302 days  44 years 270 days
1867  5                                 m       50 years 168 days  +4 years 296 days      -21 years 142 days    71 years 310 days  45 years 237 days
1863  6                                 m       79 years 28 days   +36 years 65 days      +6 years 25 days      73 years 3 days    42 years 328 days
1857  6                                 m       50 years 15 days   +5 years 172 days      -23 years 193 days    73 years 208 days  44 years 208 days
1857  6                                 m       55 years 130 days  +10 years 287 days     -18 years 78 days     73 years 208 days  44 years 208 days
1855  5                                 f       65 years 14 days   +17 years 314 days     -8 years 219 days     73 years 233 days  47 years 65 days
1853  5                                 m       71 years 81 days   +28 years 246 days     -4 years 152 days     75 years 233 days  42 years 200 days
1853  6                                 m       61 years 14 days   +18 years 179 days     -14 years 219 days    75 years 233 days  42 years 200 days
1853  6                                 f       55 years 22 days   +9 years 282 days      -18 years 335 days    73 years 357 days  45 years 105 days
1852  6                                 f       41 years 228 days  -4 years 38 days       -33 years 155 days    75 years 18 days   45 years 266 days
1851  7                                 f       80 years 18 days   +33 years 77 days      +4 years 29 days      75 years 354 days  46 years 306 days
1851  6                                 f       87 years 222 days  +40 years 281 days     +11 years 233 days    75 years 354 days  46 years 306 days
1850  5                                 f       42 years 132 days  -7 years 291 days      -33 years 309 days    76 years 76 days   50 years 58 days
1849  5                                 m       47 years 288 days  +9 years 245 days      -27 years 244 days    75 years 167 days  38 years 43 days
1848  7                                 f       68 years 363 days  +24 years 144 days     -5 years 220 days     74 years 218 days  44 years 219 days
1847  6                                 f       66 years 346 days  +24 years 146 days     -6 years 201 days     73 years 182 days  42 years 200 days
1843  6                                 f       56 years 177 days  +8 years 75 days       -17 years 82 days     73 years 259 days  48 years 102 days
1842  7                                 m       91 years 295 days  +46 years 336 days     +16 years 102 days    75 years 193 days  44 years 324 days
1841  7                                 f       79 years 327 days  +32 years 211 days     +2 years 269 days     77 years 58 days   47 years 116 days
```

<a href="#contents">Back to top</a>
## Tests

```console
$ go test
PASS
ok  	predict-death	0.153s
```
<a href="#contents">Back to top</a>
