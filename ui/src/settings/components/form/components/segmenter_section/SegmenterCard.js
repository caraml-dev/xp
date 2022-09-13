import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiPanel,
  EuiTextColor,
} from "@elastic/eui";

import { StatusBadge } from "components/status_badge/StatusBadge";
import { getSegmenterScope } from "services/segmenter/SegmenterScope";

import "./SegmenterCard.scss";
import {SegmenterSettings} from "./SegmenterSettings";

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
      <StatusBadge status={getSegmenterScope(scope)} />
    </>
  ) : (
    <>
      {displayName}
      <StatusBadge status={getSegmenterScope(scope)} />
    </>
  );

  return (
    <EuiPanel className="euiPanel--settingsSegmenterCard" paddingSize="none">
      <EuiFlexGroup alignItems="center" gutterSize="s">
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
              <SegmenterSettings
                id={id}
                name={name}
                variables={variables}
                selectedVariables={selectedVariables}
                buttonContent={buttonContent}
                errors={errors}
                onChangeSelectedVariables={onChangeSelectedVariables}
              />
            ) : (
              <>
                {displayName}
                <StatusBadge status={getSegmenterScope(scope)} />
              </>
            )}
          </EuiPanel>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
