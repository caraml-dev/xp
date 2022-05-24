import {
  EuiFlyout,
  EuiFlyoutBody,
  EuiFlyoutHeader,
  EuiText,
} from "@elastic/eui";

export const ConfigSectionFlyout = ({
  header,
  content,
  onClose,
  size,
  contentClass,
  textStyle,
}) => {
  return (
    <EuiFlyout ownFocus onClose={onClose} hideCloseButton size={size}>
      <EuiFlyoutHeader hasBorder style={{ maxHeight: 10, fontWeight: "bold" }}>
        {header}
      </EuiFlyoutHeader>
      <EuiFlyoutBody>
        <EuiText className={contentClass} style={textStyle}>
          <pre>{content}</pre>
        </EuiText>
      </EuiFlyoutBody>
    </EuiFlyout>
  );
};
