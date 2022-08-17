# Running Experiments

Experiments can be run using [routers](https://github.com/caraml-dev/turing/blob/main/docs/concepts.md) or independently.

Based on the segmenters enabled for the project and the experiment variables mapped to them (in the project's settings),
all the variable(s) must be provided in the fetch treatment request body. For each segmenter below, one of the specified
(group of) variables can be configured in the project settings and subsequently, included in the fetch treatment call.

| Segmenter           | Fetch Treatment Input (One of)    |
| ------------------- | --------------------------------- |
| S2ID                | `s2id`, `(latitude, longitude)`   |
| Days of the Week    | `tz` (timezone), `day_of_week`    |
| Hours of the Day    | `tz` (timezone), `hour_of_day`    |

## Running Experiments with Turing Routers

When deploying a Turing router, the experiment engine can be configured to 'Turing Experiments'. For more information, check [Turing - Creating a Router](https://github.com/caraml-dev/turing/tree/main/docs/how-to/create-a-router).

## Running Experiments with API

Experiments can be run independently using the POST endpoint (See details on the Treatment Swagger, in [Getting Started](./01_getting_started.md)), with the required segmenter values and randomization unit (which may be optional for some Switchback experiments) in the request body.
