import { useCallback, useState } from "react";

export const useModal = (closeModalRef) => {
  const [item, setItem] = useState();

  const openModal = useCallback((onSubmit) => {
    return (item) => {
      setItem(item);
      onSubmit();
    };
  }, []);

  const closeModal = useCallback(() => {
    setItem(undefined);
    closeModalRef.current();
  }, [closeModalRef]);

  return [item, openModal, closeModal];
};
