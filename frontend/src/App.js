import React, { useState } from "react";
import axios from "axios";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faPenToSquare,
  faChartColumn,
  faHeadset,
  faLink,
} from "@fortawesome/free-solid-svg-icons";

export default function App() {
  const [originalUrl, setOriginalUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post(
        "http://localhost:8080/short",
        { origin: originalUrl },
        { headers: { "Content-Type": "application/json" } }
      );

      if (response.status === 200) {
        console.log(response.data);
        const baseUrl = "http://localhost:8080/short";
        setShortUrl(`${baseUrl}/${response.data.id}`);
        setErrorMessage("");
      }
    } catch (error) {
      setErrorMessage("Failed to create short URL");
      console.error("Error:", error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-center text-indigo-700 mb-4">
          Create Short Links!
        </h1>
        <p className="text-xl text-center mb-8">
          _ is a custom short link personalization tool that enables you to
          target, engage, and drive more customers. Get started now.
        </p>

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
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
              >
                {shortUrl}
              </a>
            </p>
          )}
          {errorMessage && <p className="mt-4 text-red-600">{errorMessage}</p>}
        </form>

        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-indigo-700 mb-4">
            A short link, infinite possibilities
          </h2>
          <p className="text-xl mb-8">
            With the advanced intelligent link shortening service, you can
            customize your links and share them easily.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <FeatureCard
            icon={faPenToSquare}
            title="Custom Domains"
            description="Track audience individually for each brand, website, or client by using your own domain or subdomain for link shortening."
          />
          <FeatureCard
            icon={faChartColumn}
            title="Track Clicks"
            description="Focus your or your client's efforts on the most promising campaigns by taking actions based on comprehensive statistics."
          />
          <FeatureCard
            icon={faHeadset}
            title="Friendly Support"
            description="We really care about your success in using short links, so you always get answers."
          />
        </div>
      </div>
    </div>
  );
}

function FeatureCard({ icon, title, description }) {
  return (
    <div className="bg-white rounded-lg shadow-md p-6 text-center">
      <FontAwesomeIcon icon={icon} className="text-4xl text-indigo-700 mb-4" />
      <h3 className="text-xl font-bold text-indigo-700 mb-2">{title}</h3>
      <p>{description}</p>
    </div>
  );
}
