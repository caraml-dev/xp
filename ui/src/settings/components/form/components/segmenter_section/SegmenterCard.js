import { useEffect, useState } from "react";

import {
  EuiAccordion,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiHorizontalRule,
  EuiIcon,
  EuiPanel,
  EuiRadioGroup,
  EuiTextColor,
} from "@elastic/eui";
import isEqual from "lodash/isEqual";
import sortBy from "lodash/sortBy";

import { StatusBadge } from "components/status_badge/StatusBadge";

import "./SegmenterCard.scss";

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

const getScopeBadge = (scope) => {
  const status = {
    project: {
      label: "Project",
      color: "primary",
    },
    global: {
      label: "Global",
      color: "secondary",
    },
  };
  return status[scope];
};

export const SegmenterCard = ({
  id,
  name,
  isRequired,
  variables,
  selectedVariables,
  errors,
  scope,
  isDragging,
  isExpandable,
  onChangeSelectedVariables,
  dragHandleProps,
}) => {
  const displayName = isRequired ? `${name} * ` : `${name} `;
  const buttonContent = isDragging ? (
    <>
      <EuiTextColor color="accent">{displayName}</EuiTextColor>
      <StatusBadge status={getScopeBadge(scope)} />
    </>
  ) : (
    <>
      {displayName}
      <StatusBadge status={getScopeBadge(scope)} />
    </>
  );

  //TODO: Change to gear icon using arrowProps in EuiAccordion after updating to Eui >=v40.0.0
  return (
    <EuiPanel className="euiPanel--settingsSegmenterCard" paddingSize="none">
      <EuiFlexGroup alignItems="center" gutterSize="m">
        <EuiFlexItem
          className="euiFlex--settingsSegmenterCardHandle"
          grow={false}
          {...dragHandleProps}
          aria-label="Drag Handle">
          <EuiPanel
            color="success"
            className="euiPanel--settingsSegmenterCardHandle">
            <EuiIcon type="grab" size="m" />
          </EuiPanel>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiPanel paddingSize="s" color="ghostwhite">
            {!!isExpandable ? (
              <EuiAccordion
                id={id}
                paddingSize="xs"
                initialIsOpen={variables.length > 1}
                buttonContent={buttonContent}
                arrowDisplay="right">
                <EuiHorizontalRule margin="xs" />
                <VariablesMappingPanel
                  segmenterName={name}
                  variables={variables}
                  selectedVariables={selectedVariables}
                  errors={errors}
                  onChange={onChangeSelectedVariables}
                />
              </EuiAccordion>
            ) : (
              <>
                {displayName}
                <StatusBadge status={getScopeBadge(scope)} />
              </>
            )}
          </EuiPanel>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
