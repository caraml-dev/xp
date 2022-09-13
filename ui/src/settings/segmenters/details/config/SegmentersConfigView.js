import { Fragment } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer, EuiPageTemplate } from "@elastic/eui";

import { ConfigSection } from "components/config_section/ConfigSection";
import { SegmentersActivitySection } from "settings/segmenters/details/config/SegmentersActivitySection";
import { SegmentersConstraintSection } from "settings/segmenters/details/config/SegmentersConstraintSection";
import { SegmentersGeneralSection } from "settings/segmenters/details/config/SegmentersGeneralSection";
import { SegmentersOptionsSection } from "settings/segmenters/details/config/SegmentersOptionsSection";
import { SegmentersTreatmentReqSection } from "settings/segmenters/details/config/SegmentersTreatmentReqSection";

export const SegmentersConfigView = ({ segmenter }) => {
  const generalInfo = {
    title: "General Info",
    iconType: "apmTrace",
    children: <SegmentersGeneralSection segmenter={segmenter} />,
  };

  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <SegmentersActivitySection segmenter={segmenter} />,
  };

  const treatmentRequest = {
    title: "Treatment Request Fields",
    iconType: "beaker",
    children: (
      <SegmentersTreatmentReqSection
        treatmentRequestFields={segmenter.treatment_request_fields}
      />
    ),
  };

  const options = {
    title: "Options",
    iconType: "indexSettings",
    children: <SegmentersOptionsSection options={segmenter.options} />,
  };

  const constraints = {
    title: "Constraints",
    iconType: "fold",
    children: (
      <SegmentersConstraintSection constraints={segmenter.constraints || []} />
    ),
  };

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <ConfigSection
              title={generalInfo.title}
              iconType={generalInfo.iconType}>
              {generalInfo.children}
            </ConfigSection>
          </EuiFlexItem>
          {segmenter.scope === "project" && (
            <EuiFlexItem grow={1}>
              <ConfigSection title={activity.title} iconType={activity.iconType}>
                {activity.children}
              </ConfigSection>
            </EuiFlexItem>
          )}
        </EuiFlexGroup>
        <EuiSpacer size="l" />

        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <ConfigSection title={options.title} iconType={options.iconType}>
              {options.children}
            </ConfigSection>
          </EuiFlexItem>
          <EuiFlexItem grow={1}>
            <ConfigSection
              title={treatmentRequest.title}
              iconType={treatmentRequest.iconType}>
              {treatmentRequest.children}
            </ConfigSection>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="l" />

        {!!segmenter.constraints && segmenter.constraints.length > 0 && (
          <EuiFlexItem grow={1}>
            <ConfigSection
              title={constraints.title}
              iconType={constraints.iconType}>
              {constraints.children}
            </ConfigSection>
          </EuiFlexItem>
        )}
      </EuiPageTemplate.Section>
    </Fragment>
  );
};
