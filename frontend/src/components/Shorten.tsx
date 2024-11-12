import { Box } from "@mui/material";
import "../App.css";
import { useState } from "react";
import axios from "axios";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faLink } from "@fortawesome/free-solid-svg-icons";
import { useNavigate } from "react-router-dom";

const Shorten = () => {
  const [originalUrl, setOriginalUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post(
        "http://localhost/short",
        { origin: originalUrl },
        { headers: { "Content-Type": "application/json" } }
      );

      if (response.status === 200) {
        console.log(response.data);
        const baseUrl = "http://localhost:3000/short";
        setShortUrl(`${baseUrl}/${response.data.id}`);
        setErrorMessage("");
      }
    } catch (error) {
      setErrorMessage("Failed to create short URL");
      console.error("Error:", error);
    }
  };

  const handleClick = (e) => {
    e.preventDefault();
    const id = shortUrl.split("/").pop();
    // Open the preview route in a new tab
    window.open(`/short/${id}`, "_blank", "noopener,noreferrer");
  };

  return (
    <Box>
      <form
        onSubmit={handleSubmit}
        className="bg-white rounded-lg shadow-md p-6 mb-8"
      >
        <div className="flex flex-col sm:flex-row items-center bg-gray-100 rounded-lg p-2">
          <FontAwesomeIcon
            icon={faLink}
            className="text-gray-500 mr-2"
            style={{ margin: "0 8px" }}
          />
          <input
            type="url"
            placeholder="Paste a link to shorten it"
            value={originalUrl}
            onChange={(e) => setOriginalUrl(e.target.value)}
            className="flex-grow bg-transparent p-2 outline-none"
            required
          />
          <button
            type="submit"
            className="mt-2 sm:mt-0 w-full sm:w-auto bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition duration-300"
          >
            Shorten
          </button>
        </div>
        {shortUrl && (
          <p className="mt-4 text-lg">
            Shortened URL:{" "}
            <a
              href={shortUrl}
              onClick={handleClick}
              className="text-blue-600 hover:underline"
            >
              {shortUrl}
            </a>
          </p>
        )}
        {errorMessage && <p className="mt-4 text-red-600">{errorMessage}</p>}
      </form>
    </Box>
  );
};

export default Shorten;
