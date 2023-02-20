import React, { useContext, useRef } from "react";
import { AppContext } from "../context/AppContext.js";
import { UploadFileCard } from "../components/UploadFileCard";
import { DeleteFileCard } from "../components/DeleteFileCard.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";

export default function Manage() {
  const { changeApiEndPoint, apiEndPoint } = useContext(AppContext);
  const apiEndPointInput = useRef();
  const [distribute, setDistribute] = React.useState([]);
  const [process, setProcess] = React.useState([]);
  const [aggregate, setAggregate] = React.useState([]);

  const handleOnClick = () => {
    changeApiEndPoint(apiEndPointInput.current.value);
    console.log(apiEndPoint);
  };

  const handleGetAllBinaries = async () => {
    const files = await WebSocketServerService().getAllBinaries();
    setAggregate(files.data.response.AggregateBinaryNames);
    setProcess(files.data.response.ProcessBinaryNames);
    setDistribute(files.data.response.DistributeBinaryNames);

    console.log({ Binaries: files });
  };

  React.useEffect(() => {
    handleGetAllBinaries();
  }, []);

  return (
    <main className="flex gap-5 flex-col items-center pb-20 md:px-16">
      <h1 className="md:text-5xl text-3xl mb-8">Manage</h1>

      <UploadFileCard handleGetAllBinaries={handleGetAllBinaries} />

      <DeleteFileCard
        process={process}
        distribute={distribute}
        aggregate={aggregate}
        handleGetAllBinaries={handleGetAllBinaries}
      />

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
      </div>
    </main>
  );
}
