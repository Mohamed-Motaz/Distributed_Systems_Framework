import axios from "axios";

export const urlBuilder = (subPath) => {
  return `http://localhost:3001/${subPath}`;
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
