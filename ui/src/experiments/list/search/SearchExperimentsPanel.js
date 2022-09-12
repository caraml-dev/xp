import React from "react";

import { EuiButton, EuiFlyout, EuiFlyoutFooter } from "@elastic/eui";

import { SegmenterContextProvider } from "providers/segmenters/context";

import SearchExperimentsFilters from "./SearchExperimentsFilters";

import "./SearchExperimentsPanel.scss";

const SearchExperimentsPanel = ({ onChange, onClose, projectId }) => {
  return (
    <SegmenterContextProvider projectId={projectId}>
      <EuiFlyout
        id="experiments-search-panel"
        side="left"
        onClose={onClose}
        maxWidth={true}
        hideCloseButton={true}
        paddingSize="m"
        type="push"
      >
        <SearchExperimentsFilters onChange={onChange} />
        <EuiFlyoutFooter className="euiFlyoutFooter">
          <EuiButton onClick={() => onClose(false)} fill>
            Close
          </EuiButton>
        </EuiFlyoutFooter>
      </EuiFlyout>
    </SegmenterContextProvider>
  );
};

export default SearchExperimentsPanel;
