import React, { useContext, useRef } from "react";
import { AppContext } from "../context/AppContext.js";
import Home from "./Home.jsx";
import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import UploadFileButton from "../components/UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";

export default function Manage() {
  const { setApiEndPoint, apiEndPoint } = useContext(AppContext);
  const apiEndPointInput = useRef();
  const runCommandInput = useRef();

  const handleOnClick = () => {
    setApiEndPoint(apiEndPointInput.current.value);
    console.log(apiEndPoint);
  };

  return (
    <main className="flex flex-col items-center pb-20 md:px-16">
      <h1 className="md:text-5xl text-3xl mb-8">Manage</h1>

      <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
        <section className="w-full flex items-center">
          <h3 className="md:text-2xl text-xl w-52">API Endpoint</h3>
          <div className="w-full relative">
            <input
              className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black pr-36"
              type="text"
              ref={apiEndPointInput}
            />
            <button
              className="rounded-lg px-10 py-1.5 bg-blue-800 absolute right-1 top-1"
              onClick={handleOnClick}
            >
              Submit
            </button>
          </div>
        </section>
        {
          //TODO: Modify the frontend
          /**
           * handle add&delete binaries
           *
           *  */
        }
        <div>
          <section className="mt-6">
            <h3 className="md:text-2xl text-xl ">Run Command</h3>
            <textarea
              className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black"
              ref={runCommandInput}
            />
          </section>

          <UploadFileButton
            onChange={(e) =>
              handleUploadFile(
                e,
                BinariesType.process,
                runCommandInput.current.value
              )
            }
            title={BinariesType.process}
          />
          <UploadFileButton
            onChange={(e) =>
              handleUploadFile(
                e,
                BinariesType.Distribute,
                runCommandInput.current.value
              )
            }
            title={BinariesType.Distribute}
          />
          <UploadFileButton
            onChange={(e) =>
              handleUploadFile(
                e,
                BinariesType.aggregate,
                runCommandInput.current.value
              )
            }
            title={BinariesType.aggregate}
          />
        </div>
      </div>
    </main>
  );
}