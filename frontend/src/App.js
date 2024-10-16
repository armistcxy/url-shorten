import React, { useState } from "react";
import axios from "axios";
import { Box, Typography, Button, Input } from "@mui/material";
import "./App.css";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faPenToSquare,
  faChartColumn,
  faHeadset,
  faLink,
} from "@fortawesome/free-solid-svg-icons";

function App() {
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
    <Box className="main">
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Typography
          fontWeight="600"
          fontSize="3rem"
          color="rgb(48,42,150)"
          marginTop="20px"
        >
          Create Short Links!
        </Typography>
        <Typography
          maxWidth="800px"
          textAlign="center"
          marginBottom="40px"
          fontSize="1.5rem"
        >
          _ is a custom short link personalization tool that enables you to
          target, engage, and drive more customers. Get started now.
        </Typography>
        <Box
          component="form"
          onSubmit={handleSubmit}
          width="100%"
          maxWidth="800px"
          sx={{
            alignItems: "center",
            backgroundColor: "white",
            borderRadius: "12px",
            boxShadow: "0px 4px 10px rgba(0, 0, 0, 0.1)",
            padding: "20px ",
          }}
        >
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              padding: "10px ",
              backgroundColor: "rgb(228, 233, 248)",
              borderRadius: "12px",
            }}
          >
            <FontAwesomeIcon
              icon={faLink}
              fontSize="20px"
              color="rgb(145,148,158)"
            />
            <Input
              variant="outlined"
              placeholder="Paste a link to shorten it"
              fullWidth
              value={originalUrl}
              onChange={(e) => setOriginalUrl(e.target.value)}
              sx={{
                backgroundColor: "rgb(228, 233, 248)",
                borderRadius: "12px",
                "& .MuiOutlinedInput-root": {
                  borderRadius: "12px",
                },
                "& .MuiInputBase-input": {
                  fontSize: "1.2rem",
                },
                padding: "10px 20px",
              }}
              required
            />
            <Button
              type="submit"
              variant="contained"
              color="primary"
              sx={{
                fontSize: "1.2rem",
                marginLeft: "10px",
                padding: "10px 20px",
                borderRadius: "12px",
                backgroundColor: "#7f5fff",
                textTransform: "none",
                boxShadow: "none",
                ":hover": {
                  backgroundColor: "#6f4eff",
                },
              }}
            >
              Shorten
            </Button>
          </Box>
          <Box
            sx={{
              display: "flex",
              flexDirection: "column",
              justifyContent: "center",
            }}
          >
            {shortUrl && (
              <Typography
                variant="h6"
                color="black"
                fontSize="1.5rem"
                style={{ marginTop: "20px" }}
              >
                Shortened URL:{" "}
                <a href={shortUrl} target="_blank" rel="noopener noreferrer">
                  {shortUrl}
                </a>
              </Typography>
            )}
            {errorMessage && (
              <Typography color="error" style={{ marginTop: "20px" }}>
                {errorMessage}
              </Typography>
            )}
          </Box>
        </Box>
      </Box>
      <Box
        sx={{
          position: "fixed",
          bottom: "20px",
          left: 0,
          right: 0,
          justifyContent: "center",
          zIndex: 1000,
        }}
      >
        <Box textAlign="center">
          <Typography
            fontWeight="600"
            fontSize="2rem"
            color="rgb(48,42,150)"
            marginTop="40px"
          >
            A short link, infinite possibilites
          </Typography>
          <Typography fontSize="1.5rem" maxWidth="800px" margin="auto">
            With the advanced intelligent link shortening service, you can
            customize your links and share them easily.
          </Typography>
        </Box>
        <Box display="flex" justifyContent="space-evenly">
          <Box
            textAlign="center"
            borderRadius="20px"
            maxWidth="250px"
            padding="20px"
            height="200px"
            display="flex"
            flexDirection="column"
            justifyContent="center"
            alignItems="center"
          >
            <FontAwesomeIcon
              icon={faPenToSquare}
              color="rgb(48,42,150)"
              fontSize="28px"
            />
            <Typography
              fontWeight="600"
              fontSize="1.5rem"
              color="rgb(48,42,150)"
            >
              Custom Domains
            </Typography>
            <Typography>
              Track audience individually for each brand, website, or client by
              using your own domain or subdomain for link shortening.
            </Typography>
          </Box>

          <Box
            textAlign="center"
            borderRadius="20px"
            maxWidth="250px"
            padding="20px"
            height="200px"
            display="flex"
            flexDirection="column"
            justifyContent="center"
            alignItems="center"
          >
            <FontAwesomeIcon
              icon={faChartColumn}
              color="rgb(48,42,150)"
              fontSize="28px"
            />
            <Typography
              fontWeight="600"
              fontSize="1.5rem"
              color="rgb(48,42,150)"
            >
              Track Clicks
            </Typography>
            <Typography>
              Focus your or your client's efforts on the most promising
              campaigns by taking actions based on comprehensive statistics.
            </Typography>
          </Box>
          <Box
            textAlign="center"
            borderRadius="20px"
            maxWidth="250px"
            padding="20px"
            height="200px"
            display="flex"
            flexDirection="column"
            justifyContent="center"
            alignItems="center"
          >
            <FontAwesomeIcon
              icon={faHeadset}
              color="rgb(48,42,150)"
              fontSize="28px"
            />
            <Typography
              fontWeight="600"
              fontSize="1.5rem"
              color="rgb(48,42,150)"
            >
              Friendly Support
            </Typography>
            <Typography>
              We really care about your success in using short links, so you
              always get answers.
            </Typography>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}

export default App;
