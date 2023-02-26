import { ProgressIndicator } from "@fluentui/react";
import { Progress, Tooltip } from "flowbite-react";
import React, { useContext, useEffect, useState } from "react";
import Loading from "../components/Loading.jsx";
import StatusCard from "../components/StatusCard.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { AppContext } from "./../context/AppContext";


export default function Status() {
  const [jobs, setJobs] = useState(null);
  const [loading, setLoading] = useState(true);

  const { TriggerAlert } = useContext(AppContext);

  const getJobsProgress = async () => {
    const jobProgress = await WebSocketServerService().getJobProgress();
    if (!jobProgress?.data?.success) {
      TriggerAlert(
        jobProgress?.data?.response ??
        "Unable to establish the communication with the server"
      );
    }
    setJobs(jobProgress.data.response || []);

    // setJobs([responseExample(true,"Processing"), responseExample(true,"Free"), responseExample(true,"Unresponsive")])
  };

  useEffect(() => {
    getJobsProgress();

    const intervalCalling = setInterval(async () => {
      //console.log("getJobsProgress() : Start...");
      await getJobsProgress();
      //console.log("getJobsProgress() : Done");
    }, 5000);

    return () => {
      clearInterval(intervalCalling);
    };
  }, []);

  return (
    <main className="flex flex-col items-center pb-20 md:px-6">
      <h1 className="md:text-5xl text-3xl mb-8">Status</h1>

      {loading && jobs === null ? (
        <Loading />
      ) : jobs.length ? (
        <section className="w-full grid grid-cols-12">
          {jobs.map((job, index) => (
            <StatusCard
              key={index}
              // key={job.response.Progress[0].JobId}
              job={job}
            />
          ))}
        </section>
      ) : (
        <h2 className="text-2xl self-start">No Jobs Found</h2>
      )}
    </main>
  );
}
