import {
  EuiAccordion,
  EuiButtonIcon,
  EuiFormRow,
  EuiHorizontalRule,
  EuiPanel,
  EuiRadioGroup,
} from "@elastic/eui";
import {useEffect, useState} from "react";
import isEqual from "lodash/isEqual";
import sortBy from "lodash/sortBy";

const VariablesMappingPanel = ({
  segmenterName,
  variables,
  selectedVariables,
  errors,
  onChange,
}) => {
  const options = variables.map((names, idx) => ({
    id: `${segmenterName}-${idx}`,
    label: names.join(", "),
    names: names, // Store the names as is, for selection comparison
  }));
  const [selectedOption, setSelectedOption] = useState();

  useEffect(() => {
    const selected = options.find((e) =>
      isEqual(sortBy(e.names), sortBy(selectedVariables))
    );
    if (!!selected && selectedOption !== selected.id) {
      setSelectedOption(selected.id);
    }
  }, [options, selectedOption, selectedVariables]);

  const onChangeSelectedVariables = (id) => {
    const item = options.find((e) => e.id === id);
    onChange(!!item ? item.names : []);
  };

  return (
    <EuiPanel paddingSize="none" color="ghostwhite">
      <EuiFormRow fullWidth isInvalid={!!errors} error={errors}>
        <EuiRadioGroup
          options={options}
          idSelected={selectedOption}
          onChange={onChangeSelectedVariables}
          legend={{
            children: <span>Experiment Variable(s)</span>,
          }}
        />
      </EuiFormRow>
    </EuiPanel>
  );
};

export const SegmenterSettings = ({
  id,
  name,
  variables,
  selectedVariables,
  buttonContent,
  errors,
  onChangeSelectedVariables,
}) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <EuiAccordion
      id={id}
      forceState={isOpen ? "open" : "closed"}
      paddingSize="xs"
      initialIsOpen={variables.length > 1}
      buttonContent={buttonContent}
      arrowDisplay="none"
      extraAction={
        <EuiButtonIcon size="s" iconType={"gear"} color={"text"} onClick={() => setIsOpen(!isOpen)}/>
      }
    >
      <EuiHorizontalRule margin="xs" />
      <VariablesMappingPanel
        segmenterName={name}
        variables={variables}
        selectedVariables={selectedVariables}
        errors={errors}
        onChange={onChangeSelectedVariables}
      />
    </EuiAccordion>
  );
};
