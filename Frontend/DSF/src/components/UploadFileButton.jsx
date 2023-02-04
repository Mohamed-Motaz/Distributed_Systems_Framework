import { Button } from "flowbite-react";
import React, { useRef } from "react";
import { WebSocketServerService } from "./../services/WebSocketServerService";
import { BinariesType } from "./../services/ServiceTypes/WebSocketServiceTypes";

export default function UploadFileButton(props) {
  const { title, onChange } = props;
  const hiddenFileInput = useRef(null);

  function handleClick() {
    hiddenFileInput.current.click();
  }

  return (
    <button className="m-2">
      <Button onClick={handleClick}>{title}</Button>
      <input
        type="file"
        ref={hiddenFileInput}
        onChange={onChange}
        className="hidden"
      />
    </button>
  );
}
