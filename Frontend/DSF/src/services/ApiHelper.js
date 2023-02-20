import axios from "axios";

export const urlBuilder = (subPath) => {
  const baseUrl = localStorage.getItem('apiEndPoint')
  return `http://${baseUrl}/${subPath}`;
};

// export const ApiHelper = () => {
//   func = async (mediaType) => {
//     let response;

//     await axios
//       .get(
//         `https://api.themoviedb.org/3/trending/${mediaType}/week?api_key=${API_KEY}`
//       )
//       .then((value) => (response = value))
//       .catch((error) => console.log("ERROR: " + error));

//     return response;
//   };
// };
