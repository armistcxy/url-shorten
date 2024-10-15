import React, { useState } from "react";
import axios from "axios";

function App() {
  const [originalUrl, setOriginalUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post(
        "http://localhost:8080/short",
        {
          origin: originalUrl,
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      if (response.status === 200) {
        setShortUrl(response.data.shortUrl);
        setErrorMessage("");
      }
    } catch (error) {
      setErrorMessage("Failed to create short URL");
      console.error("Error:", error);
    }
  };

  return (
    <div className="App">
      <h1>URL Shortener</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Original URL:</label>
          <input
            type="text"
            value={originalUrl}
            onChange={(e) => setOriginalUrl(e.target.value)}
            required
          />
        </div>
        <button type="submit">Shorten URL</button>
      </form>
      {shortUrl && (
        <div>
          <h2>Shortened URL:</h2>
          <a href={shortUrl} target="_blank" rel="noopener noreferrer">
            {shortUrl}
          </a>
        </div>
      )}
      {errorMessage && <p style={{ color: "red" }}>{errorMessage}</p>}
    </div>
  );
}

export default App;
