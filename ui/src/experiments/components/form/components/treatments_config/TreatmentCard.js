import React, { Fragment, useContext, useEffect, useState } from "react";

import {
  EuiButtonIcon,
  EuiComboBox,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiLoadingChart,
  EuiPanel,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";

import { useXpApi } from "hooks/useXpApi";
import TreatmentContext from "providers/treatment/context";

import { TreatmentPanel } from "./TreatmentPanel";

export const TreatmentCard = ({
  treatment,
  onUpdateTreatment,
  onDelete,
  projectId,
  errors = {},
}) => {
  const { isLoaded, treatments } = useContext(TreatmentContext);
  const [treatmentId, setTreatmentId] = useState();
  const [hasNewResponse, setHasNewResponse] = useState(false);

  const treatmentSelectionOptions = treatments.map((treatment) => {
    return {
      label: treatment.name,
      id: treatment.id,
    };
  });

  const [
    { data: treatmentDetails, isLoaded: isAPILoaded },
    fetchtreatmentDetails,
  ] = useXpApi(
    `/projects/${projectId}/treatments/${treatmentId}`,
    {},
    {},
    false
  );

  const onChange = (selected) => {
    // If id is not present, it will be set to undefined, but will not trigger useEffect
    setTreatmentId(selected[0]?.id);
    onUpdateTreatment(
      treatment.uuid,
      "template",
      selected.length > 0 ? selected[0] : ""
    );

    if (selected.length === 0 || !selected[0].id) {
      // Either custom treatment or the selection is cleared
      onUpdateTreatment(treatment.uuid, "name", "");
      onUpdateTreatment(treatment.uuid, "traffic", 0);
      onUpdateTreatment(treatment.uuid, "configuration", "");
    }
  };

  // Fetch treatment details every time there is a selection of treatment with id
  useEffect(() => {
    if (!!treatmentId) {
      fetchtreatmentDetails();
      setHasNewResponse(true);
    }
  }, [treatmentId, fetchtreatmentDetails]);

  // Populate Treatment object so that it will reflect on UI
  useEffect(() => {
    if (hasNewResponse && isAPILoaded) {
      onUpdateTreatment(treatment.uuid, "name", treatmentDetails.data.name);
      onUpdateTreatment(
        treatment.uuid,
        "configuration",
        JSON.stringify(treatmentDetails.data.configuration)
      );
      setHasNewResponse(false);
    }
  }, [
    treatmentDetails,
    treatment.uuid,
    onUpdateTreatment,
    hasNewResponse,
    isAPILoaded,
  ]);

  return (
    <Fragment>
      {isLoaded ? (
        <EuiPanel>
          <EuiFlexGroup gutterSize="none" direction="column">
            <EuiFlexGroup justifyContent="flexEnd" alignItems="flexEnd">
              <EuiFlexItem grow={false}>
                {!!onDelete ? (
                  <EuiButtonIcon
                    iconType="cross"
                    onClick={onDelete}
                    aria-label="delete-treatment"
                  />
                ) : (
                  <EuiIcon type="empty" size="l" />
                )}
              </EuiFlexItem>
            </EuiFlexGroup>
            <EuiFormRow fullWidth label="Template">
              <EuiComboBox
                placeholder="Copy from Pre-configured Treatment"
                isDisabled={
                  !treatmentSelectionOptions ||
                  treatmentSelectionOptions.length === 0
                }
                fullWidth={true}
                singleSelection={{ asPlainText: true }}
                options={treatmentSelectionOptions}
                onChange={onChange}
                selectedOptions={
                  !!treatment.template ? [treatment.template] : []
                }
              />
            </EuiFormRow>
            <EuiSpacer size="m" />
            <TreatmentPanel
              errors={errors}
              treatment={treatment}
              onUpdateTreatment={onUpdateTreatment}
            />
          </EuiFlexGroup>
        </EuiPanel>
      ) : (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      )}
    </Fragment>
  );
};
