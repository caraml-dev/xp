import React, { useContext, useEffect, useState } from "react";

import {
  EuiButton,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFlyout,
  EuiFlyoutBody,
  EuiFlyoutFooter,
  EuiFlyoutHeader,
  EuiFormRow,
  EuiText,
  EuiTextArea,
  EuiTitle,
} from "@elastic/eui";
import { FormContext, FormLabelWithToolTip, addToast } from "@gojek/mlp-ui";

import { useXpApi } from "hooks/useXpApi";
import { ValidateEntityRequest } from "services/validate/ValidateEntityRequest";
import schema from "settings/components/form/validation/schema";
import { PlaygroundRadioGroup } from "settings/components/playground_flyout/PlaygroundRadioGroup";
import { getValidationOptions } from "settings/components/playground_flyout/typeOptions";

var JSONbig = require("json-bigint");

const PlaygroundFlyout = ({ onClose }) => {
  const { data: settings } = useContext(FormContext);

  const [selectedValidationType, setSelectedValidationType] = useState(null);
  const [validationData, setValidationData] = useState("");
  const [inputErrors, setInputErrors] = useState([]);

  const [submissionResponse, submitValidation] = useXpApi(
    "/validate",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => {
    let validationRequest = ValidateEntityRequest.fromSettings(settings);
    setInputErrors([]); // reset the errors
    try {
      schema[4].validateSync(validationData, { abortEarly: false });
    } catch (e) {
      setInputErrors([["Invalid input: enter a JSON object"]]);
      return;
    }
    validationRequest.data = JSONbig.parse(validationData);
    return submitValidation({
      body: validationRequest.setupRequest(selectedValidationType).stringify(),
    }).promise;
  };
  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-update-settings",
        title: "Validation success!",
        color: "success",
        iconType: "check",
      });
    }
  }, [submissionResponse]);
  useEffect(() => {
    const enabledOptions = getValidationOptions(settings).filter(
      (e) => !e.disabled && e.id === selectedValidationType
    );
    if (enabledOptions.length === 0) {
      setSelectedValidationType(null);
    }
  }, [settings, selectedValidationType]);

  return (
    <EuiFlyout
      onClose={onClose}
      size="313px"
      maxWidth={true}
      paddingSize="m"
      type="push"
    >
      <EuiFlyoutHeader hasBorder>
        <EuiFlexItem>
          <EuiTitle size="s">
            <h4>Playground</h4>
          </EuiTitle>
        </EuiFlexItem>
      </EuiFlyoutHeader>
      <EuiFlyoutBody>
        <EuiText>
          <EuiFlexItem>
            <PlaygroundRadioGroup
              selectedValidationType={selectedValidationType}
              setSelectedValidationType={setSelectedValidationType}
            />
          </EuiFlexItem>
          <br />
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Sample Data"
                  content="Sample data that you would like to validate. This may either be a request payload (to be
                  validated with an external validation endpoint), or a treatment configuration (to be validated against
                  treatment validation rules)."
                />
              }
              display="row"
              isInvalid={inputErrors.length !== 0}
              error={[inputErrors]}>
              <EuiTextArea
                fullWidth
                placeholder={
                  !!selectedValidationType
                    ? getValidationOptions(settings).find(
                        (e) => e.id === selectedValidationType
                      )?.placeholderText
                    : `Click on one of the validation types above to begin.`
                }
                value={!selectedValidationType ? "" : validationData}
                onChange={(e) =>
                  !!selectedValidationType
                    ? setValidationData(e.target.value)
                    : null
                }
                isInvalid={inputErrors.length !== 0}
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiText>
      </EuiFlyoutBody>
      <EuiFlyoutFooter>
        <EuiFlexGroup justifyContent="flexEnd">
          <EuiFlexItem grow={false}>
            <EuiButton
              onClick={onSubmit}
              isDisabled={!selectedValidationType}
              fill>
              Validate
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlyoutFooter>
    </EuiFlyout>
  );
};

export default PlaygroundFlyout;
