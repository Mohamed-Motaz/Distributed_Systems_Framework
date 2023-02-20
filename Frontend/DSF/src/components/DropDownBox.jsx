import { Dropdown } from "flowbite-react";
import React, { useState } from "react";

export default function DropDownBox(props) {
  const { title, files, selectedFile, setSelectedFile } = props;

  function handleCountrySelect(e) {
    console.log("Selected file", e.target.value);
    const f = e.target.value;
    setSelectedFile(f);
  }

  console.log({ files });

  // return (
  //   <div className="dropDownBox">
  //     <h1>{title}</h1>

  //     <div className="Container">
  //       <h2>{selectedFile}</h2>
  //       <select
  //         name="process"
  //         onChange={(e) => handleCountrySelect(e)}
  //         value={selectedFile}
  //       >
  //         <option value="">{`Select ${title} File`}</option>
  //         {files?.map((file, key) => (
  //           <option key={key} value={file}>
  //             {file}
  //           </option>
  //         ))}
  //       </select>
  //     </div>
  //   </div>
  // );

  return <Dropdown onChange={()=>{}} className='dark' label="Menu" style={{ background: '#1744e1' }}>
    <Dropdown.Item className='justify-center'>
      Item 1
    </Dropdown.Item>
    <Dropdown.Item className='justify-center'>
      Item 2
    </Dropdown.Item>
  </Dropdown>
}
