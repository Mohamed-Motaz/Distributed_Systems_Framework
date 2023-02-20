import { Button } from "flowbite-react";
import React, { useRef } from "react";
import { WebSocketServerService } from "./../services/WebSocketServerService";
import { BinariesType } from "./../services/ServiceTypes/WebSocketServiceTypes";

export default function UploadFileButton(props) {
  const { title, onChange, inputType } = props;
  const hiddenFileInput = useRef(null);

  const [optionalText, setOptionalText] = React.useState(
    "Choose optional file"
  );

  function handleClick() {
    hiddenFileInput.current.click();
  }

  const handleChange = (e) => {
    onChange(e);
    setOptionalText(e.target.files[0].name);
  };

  return (
    <div className="flex gap-2 w-full justify-center items-center mt-5">
      <button className="m-2">
        <Button onClick={handleClick}>{title}</Button>
        <input
          type={inputType ?? "file"}
          ref={hiddenFileInput}
          onChange={(e) => handleChange(e)}
          className="hidden"
          accept=".zip,.rar,.7zip,.tar"
        />
      </button>
      <p>{optionalText}</p>
    </div>
  );
}
