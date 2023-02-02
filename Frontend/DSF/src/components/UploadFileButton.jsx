import { Button } from 'flowbite-react';
import React, { useRef } from 'react'

export default function UploadFileButton(props) {
    const { title } = props;
    const hiddenFileInput = useRef(null);

    function handleClick() {
        hiddenFileInput.current.click();
    };

    function handleChange(event) {
        const fileUploaded = event.target.files[0];
        console.log(fileUploaded);
    };

    return <button className='m-2'>
        <Button onClick={handleClick}>
            {title}
        </Button>
        <input
            type="file"
            ref={hiddenFileInput}
            onChange={handleChange}
            className='hidden'
        />
    </button>
}

