import { ProgressIndicator } from "@fluentui/react";
import { Progress, Tooltip } from "flowbite-react";
import React, { useContext, useEffect, useState } from "react";
import Loading from "../components/Loading.jsx";
import StatusCard from "../components/StatusCard.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { AppContext } from "./../context/AppContext";

function jobExample(jobStatus) {
  return {
    MasterId: "idmd9303-kdk9303-eke-2993iop",
    JobId: "idmd9303-kdk9303-eke-2993iop",
    ClientId: "idmd9303-kdk9303-eke-2993iop",
    Progress: 34.56,
    Status: jobStatus,
    WorkersTasks: [
      {
        WorkerId: "idmd9303-kdk9303-eke-2993iop",
        CurrentTaskContent: "lastTask.txt",
        FinishedTaskContent: [
          "FirstTask.txt",
          "SecondTask.txt",
          "ThirdTask.txt",
          "FourthTask.txt",
        ],
      },
      {
        WorkerId: "idmd9303-kdk9303-eke-2993iop",
        CurrentTaskContent: "lastTask.txt",
        FinishedTaskContent: [
          "FirstTask.txt",
          "SecondTask.txt",
          "ThirdTask.txt",
          "FourthTask.txt",
        ],
      },
      {
        WorkerId: "idmd9303-kdk9303-eke-2993iop",
        CurrentTaskContent: "lastTask.txt",
        FinishedTaskContent: [
          "FirstTask.txt",
          "SecondTask.txt",
          "ThirdTask.txt",
          "FourthTask.txt",
        ],
      },
    ],
  };
}

export default function Status(props) {
  const { jobs } = props;
  const [loading, setLoading] = useState(true);

  const { TriggerAlert } = useContext(AppContext);

  return (
    <main className="flex flex-col items-center pb-20 md:px-6">
      <h1 className="md:text-5xl text-3xl mb-8">Status</h1>

      {loading && jobs === null ? (
        <Loading />
      ) : jobs.length ? (
        <section className="w-full grid grid-cols-12">
          {jobs.map((job) => (
            <StatusCard key={job.MasterId} job={job} />
          ))}
        </section>
      ) : (
        <h2 className="text-2xl self-start">No Jobs Found</h2>
      )}
    </main>
  );
}
