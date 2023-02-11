import React, { useState } from "react";

export default function DropDownBox(props) {
  const { title, files, selectedFile, setSelectedFile } = props;

  function handleCountrySelect(e) {
    console.log("Selected file", e.target.value);
    const f = e.target.value;
    setSelectedFile(f);
  }

  console.log({ files });

  return (
    <div className="dropDownBox">
      <h1>{title}</h1>

      <div className="Container">
        <select
          name="process"
          onChange={(e) => handleCountrySelect(e)}
          value={selectedFile}
        >
          <option value="">{`Select ${title} File`}</option>
          {files?.map((file, key) => (
            <option key={key} value={file}>
              {file}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}
