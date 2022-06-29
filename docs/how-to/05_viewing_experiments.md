# Viewing Experiments

Once the experiment has been created, you will be able to view the experiment's configuration on the landing page.

![View Experiment](../assets/05_view_experiment_landing.png)

## Navigate to Experiment Details

1. Click on the row that contains the experiment.
2. You will now be able to see the Experiment Details View.
![View Experiment Details](../assets/05_view_experiment_detail.png)
At the top row, you will be able to see your experiment name and a badge that indicates the status of experiment

| Status       | Description                                          | Badge                                                                                   |
| ------------ | ---------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Running      | Experiment is active and currently running           | ![View Experiment Status Running](../assets/05_view_experiment_status_running.png)      |
| Scheduled    | Experiment is active and start time is in the future | ![View Experiment Status Scheduled](../assets/05_view_experiment_status_scheduled.png)  |
| Completed    | Experiment is active and end time is the past       | ![View Experiment Status Completed](../assets/05_view_experiment_status_completed.png)  |
| Deactivated  | Experiment is inactive                               | ![View Experiment Status Inactive](../assets/05_view_experiment_status_deactivated.png) |

### Configuration

The Configuration tab displays the selected experiment's details. These values are configured from creating or editing an experiment.

1. General Info: General settings of the experiment.
2. Activity: Activity details of experiment.
3. Segment: Segmenters of experiment.
4. Treatments: Treatments Configurations for 1 or more registered Treatment(s).

### Searching

The UI supports two types of search - Basic and Advanced. Advanced search options enable filtering the experiments by different attributes.

## Basic Search

1. In the search panel, enter the experiment name or description to filter by.
   ![View Experiment Search Simple](../assets/05_view_experiment_search_simple.png)

## Advanced Search

1. Click "Search Options", this will open up the Filters Panel.
   ![View Experiment Search Filter](../assets/05_view_experiment_search_filter.png)

2. In the Filters Panel, select the respective filters to apply. A "Filtered" badge will be shown beside experiment name to indicate that the experiments are filtered.
   ![View Experiment Search Filter](../assets/05_view_experiment_search_filtered.png)

### History

When an experiment is modified (edited / activated / deactivated) its existing configurations are saved as a historical version. All versions can be viewed from the **History** tab of the Experiment Detail view.

![View Experiment History](../assets/05_view_experiment_history.png)

The versions are ordered in the descending order of creation (the most recent version appearing on top). The Created and Updated dates of the version symbolize the duration that the configuration was applied in the experiment. Selecting a row opens the details of the version.

![View Experiment History](../assets/05_view_experiment_historical_version.png)
