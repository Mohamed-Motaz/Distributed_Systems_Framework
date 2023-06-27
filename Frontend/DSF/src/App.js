import React, { useContext, useEffect, useState } from "react";
import { RouterProvider, createBrowserRouter } from "react-router-dom";
import useWebSocket from "react-use-websocket";
import { AppContext } from "../src/context/AppContext";
import "./App.css";
import AboutUs from "./Pages/AboutUs.jsx";
import FinishedJobs from "./Pages/FinishedJobs.jsx";
import HowTo from "./Pages/HowTo.jsx";
import Landing from "./Pages/Landing.jsx";
import Manage from "./Pages/Manage.jsx";
import NotFound from "./Pages/NotFound";
import Status from "./Pages/Status.jsx";
import SubmitJob from "./Pages/SubmitJob.jsx";
import RootLayout from "./components/RootLayout";
import useAlert from "./helpers/useAlert";

export default function App() {
  const [isFirst, setIsFirst] = useState(true);

  const { clientId, apiEndPoint } = useContext(AppContext);
  const WS_URL = `ws://${apiEndPoint}/openWS/${clientId}`;

  const [AlertComponent, TriggerAlert] = useAlert();

  const [isSuccess, setIsSuccess] = React.useState(false);

  const [binaries, setBinaries] = useState({
    process: [],
    aggregate: [],
    distribute: [],
  });

  const [finishedJobIds, setFinishedJobIds] = useState(null);
  const [systemProgress, setSystemProgress] = useState(null);

  const setAllBinaries = async (files) => {
    const { AggregateBinaryNames, ProcessBinaryNames, DistributeBinaryNames } =
      files;
    setBinaries({
      process: ProcessBinaryNames,
      aggregate: AggregateBinaryNames,
      distribute: DistributeBinaryNames,
    });
  };

  //getAllBinaries and getSystemProgress and finishedJobsIds
  console.log("API Endpoint =======>>>> ", apiEndPoint);
  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    shouldReconnect: (closeEvent) => true,
    // onClose: () => {
    //   console.log(
    //     "WebSocket connection closed, it will be re-established in a second"
    //   );
    //   setTimeout(wsClient, 1000);
    // },
    onMessage: (e) => {
      const wsResponse = JSON.parse(e.data);
      console.log({ e: wsResponse });
      if (wsResponse.msgType === "systemBinaries") {
        if (wsResponse.response.success) {
          setAllBinaries(wsResponse.response.response);
        } else {
          TriggerAlert(
            e?.data?.response.response ?? "Unable to get system binaries"
          );
        }
      } else if (wsResponse.msgType === "finishedJobsIds") {
        if (wsResponse.response.success) {
          setFinishedJobIds(e?.data?.response.response || []);
        } else {
          TriggerAlert(
            wsResponse.response.response ?? "Unable to get finished job IDs"
          );
        }
      } else if (wsResponse.msgType === "systemProgress") {
        if (wsResponse.response.success) {
          setSystemProgress(wsResponse.response.response || []);
        } else {
          TriggerAlert(
            wsResponse.response.response ?? "Unable to get systemProgress"
          );
        }
      } else if (wsResponse.msgType === "finishedJob") {
        if (wsResponse.response.success) {
          TriggerAlert(
            `The job with id: ${wsResponse.response.response.jobId} is finished`,
            () => { }
          ); // implement download logic
        } else {
          TriggerAlert(
            wsResponse.response.response ?? "Unable to get finished job"
          );
        }
      } else if (wsResponse.msgType === "jobRequest") {
        if (wsResponse.response.success) {
          TriggerAlert(
            `The job with id: ${wsResponse.response.response.JobId} is finished`,
            () => { }
          ); // implement download logic
        } else {
          TriggerAlert(
            wsResponse.response.response ?? `Job has an error: ${wsResponse.response.response}`
          );
        }
      }
      else {
        TriggerAlert(`Unexpected message type: ${wsResponse.msgType}`);
      }
    },
  });

  const HOME_ROUTE = createBrowserRouter([
    {
      path: "/",
      element: <RootLayout />,
      children: [
        { index: true, element: <Landing /> },
        { path: "/how-to", element: <HowTo /> },
        { path: "/manage", element: <Manage binaries={binaries} /> },
        {
          path: "/submit-job",
          element: <SubmitJob binaries={binaries} wsClient={wsClient} />,
        },
        { path: "/status", element: <Status jobs={systemProgress} /> },
        {
          path: "/finished-jobs",
          element: <FinishedJobs jobs={finishedJobIds} />,
        },
        { path: "/about-us", element: <AboutUs /> },
        // {
        //   path: '/movies', element: <Outlet></Outlet>,
        //   children: [
        //     { index: true, element: <Movies /> },
        //     { path: 'details/:type/:id', element: <ItemDetails /> },
        //   ]
        // },
        { path: "*", element: <NotFound /> },
      ],
    },
  ]);

  useEffect(() => {
    const result = localStorage.getItem("isFirst");
    if (result) {
      setIsFirst(false);
      localStorage.setItem("isFirst", "False");
    } else {
      localStorage.setItem("isFirst", "True");
      setIsFirst(true);
    }

    console.log("====================================");
    console.log("DID APP MOUNT");
    console.log("====================================");
  }, []);

  return (
    <main>
      <AlertComponent success={isSuccess} />

      <RouterProvider
        router={
          localStorage.getItem("apiEndPoint")
            ? HOME_ROUTE
            : createBrowserRouter([
              {
                path: "/",
                element: (
                  <div className="dark pt-28 px-8">
                    <Landing />
                  </div>
                ),
              },
            ])
        }
      />
    </main>
  );
}
