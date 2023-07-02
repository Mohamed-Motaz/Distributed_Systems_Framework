import { Progress, Tooltip } from "flowbite-react";
import React, { useState } from "react";
import WorkersModal from "./WorkersModal.jsx";

export default function StatusCard({ job }) {
  console.log(job.error ? "Error Found" : "Success");

  const [isCopied, setIsCopied] = useState(false);
  const [timer, setTimer] = useState(null);

  function copyToClipboard() {
    clearTimeout(timer);
    navigator.clipboard.writeText(job.JobId);
    setIsCopied(true);
    setTimer(setTimeout(() => setIsCopied(false), 1000));
  }

  return job.Status === "Processing"
    ? ProgressingMaster(job, isCopied, copyToClipboard)
    : job.Status === "Free"
    ? FreeMaster(job, isCopied, copyToClipboard)
    : UnresponsiveMaster(job, isCopied, copyToClipboard);
}

function ProgressingMaster(job, isCopied, copyToClipboard) {
  return (
    <div className="xl:col-span-4 md:col-span-6 col-span-12 px-3 mb-6 flex">
      <div className="rounded-lg p-8 pb-4 self-stretch w-full gap-2 bg-green-900">
        <div className="mb-8 flex justify-center items-center gap-2">
          <button
            className="rounded-lg bg-black border-blue-800 border-2"
            onClick={copyToClipboard}
          >
            <Tooltip content={isCopied ? "Copied To Cilpboard" : job.MasterId}>
              <h2 className="text-xl text-center px-6 py-2">
                MASTER SERVER ID
              </h2>
            </Tooltip>
          </button>
        </div>

        <div className="mb-3 flex items-center gap-2">
          <div className="w-24 mr-4">Job ID:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.JobId}</p>
          </div>
        </div>
        <div className="mb-3 flex items-center gap-2">
          <div className="w-24">Process:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.ProcessBinaryName}</p>
          </div>
        </div>
        <div className="mb-3 flex items-center gap-2">
          <div className="w-24">Distribute:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.DistributeBinaryName}</p>
          </div>
        </div>
        <div className="mb-3 flex items-center gap-2">
          <div className="w-24">Aggregate:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.AggregateBinaryName}</p>
          </div>
        </div>
        <div className="mb-3 flex items-center gap-2">
          <div className="w-24">Time Assigned:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.TimeAssigned}</p>
          </div>
        </div>
        <div className="mb-3 flex items-center gap-2">
          <div className="w-24">Created At:</div>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{job.CreatedAt}</p>
          </div>
        </div>
        <div className="mt-6 mb-3">
          <Progress
            progress={job.Progress}
            color="green"
            label={`Job Progress: ${job.Status}`}
            labelPosition="outside"
            labelProgress={true}
          />
        </div>

        <WorkersModal workers={job.WorkersTasks} />
      </div>
    </div>
  );
}

function FreeMaster(job, isCopied, copyToClipboard) {
  return (
    <div className="xl:col-span-4 md:col-span-6 col-span-12 px-3 mb-6 flex">
      <div className="rounded-lg p-8 self-stretch w-full gap-2 bg-gray-700">
        <div className="mb-8 flex justify-center items-center gap-2">
          <button
            className="rounded-lg bg-black border-blue-800 border-2"
            onClick={copyToClipboard}
          >
            <Tooltip content={isCopied ? "Copied To Cilpboard" : job.MasterId}>
              <h2 className="text-xl text-center px-6 py-2">
                MASTER SERVER ID
              </h2>
            </Tooltip>
          </button>
        </div>
        <div className="text-center text-2xl pt-[20%]">Free Master</div>
      </div>
    </div>
  );
}

function UnresponsiveMaster(job, isCopied, copyToClipboard) {
  return (
    <div className="xl:col-span-4 md:col-span-6 col-span-12 px-3 mb-6 flex">
      <div className="rounded-lg p-8 self-stretch w-full gap-2 bg-red-900">
        <div className="mb-8 flex flex-col justify-center items-center gap-2">
          <button
            className="rounded-lg bg-black border-blue-800 border-2"
            onClick={copyToClipboard}
          >
            <Tooltip content={isCopied ? "Copied To Cilpboard" : job.MasterId}>
              <h2 className="text-xl text-center px-6 py-2">
                MASTER SERVER ID
              </h2>
            </Tooltip>
          </button>
        </div>
        <div className="text-center text-2xl pt-[20%]">Unresponsive Master</div>
      </div>
    </div>
  );
}
