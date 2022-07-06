import "experiments/list/search/SearchExperimentsPanel.scss";

import React from "react";

import { EuiButton, EuiFlyout, EuiFlyoutFooter } from "@elastic/eui";

import SearchExperimentsFilters from "experiments/list/search/SearchExperimentsFilters";
import { SegmenterContextProvider } from "providers/segmenters/context";

const SearchExperimentsPanel = ({ onChange, onClose, projectId }) => {
  return (
    <SegmenterContextProvider projectId={projectId}>
      <EuiFlyout
        id="experiments-search-panel"
        className="searchPanelFlyout--left"
        onClose={onClose}
        size="s"
        maxWidth={true}
        hideCloseButton={true}
        paddingSize="m">
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
