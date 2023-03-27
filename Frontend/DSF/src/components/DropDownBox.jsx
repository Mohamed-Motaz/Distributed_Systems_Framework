import { Dropdown } from "flowbite-react";
import React, { useState } from "react";

export default function DropDownBox(props) {
  const { title, files, selectedFile, setSelectedFile } = props;

  function handlechange(e) {
    console.log("Selected file", e.target.value);
    const f = e.target.value;
    setSelectedFile(f === "Choose file..." ? undefined : f);
    if (f !== "Choose file...") {
      sessionStorage.setItem(title, f);
    }
  }

  console.log({ files, selectedFile });

  return (
    <div className="mb-4">
      <label
        htmlFor="countries"
        className="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
      >
        {`Select ${title} file`}
      </label>
      <select
        defaultValue={selectedFile}
        value={selectedFile}
        placeholder="choose file"
        id="countries"
        onChange={(e) => handlechange(e)}
        className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
      >
        <option>Choose file...</option>
        {files?.map((file) => (
          <option>{file}</option>
        ))}
      </select>
    </div>
  );
}
