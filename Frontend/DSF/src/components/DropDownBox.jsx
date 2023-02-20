import { Dropdown } from "flowbite-react";
import React, { useState } from "react";

export default function DropDownBox(props) {
  const { title, files, selectedFile, setSelectedFile } = props;
  const [fileName, setFileName] = React.useState(title);

  function handlechange(e) {
    console.log("Selected file", e.target.value);
    const f = e.target.value;
    setFileName(f);
  }

  console.log({ files });

  return (
    <div className="mb-4">
      <label
        for="countries"
        class="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
      >
        {`Select ${title} file`}
      </label>
      <select
        id="countries"
        class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
      >
        {files?.map((file) => (
          <option>{file}</option>
        ))}
      </select>
    </div>
  );
}
