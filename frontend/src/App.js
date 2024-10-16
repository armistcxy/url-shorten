import React, { useState } from "react";
import axios from "axios";
import {
  Box,
  IconButton,
  Typography,
  Avatar,
  Button,
  TextField,
  InputAdornment,
  Input,
} from "@mui/material";
import "./App.css";
import LinkIcon from "@mui/icons-material/Link";

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
    <Box>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          height: "400px",
          backgroundColor: "white",
        }}
      >
        <Typography fontWeight="600" fontSize="3rem" color="rgb(48,42,150)">
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
        <Box onSubmit={handleSubmit} width="100%" maxWidth="800px">
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              backgroundColor: "#f0f0f0",
              borderRadius: "12px",
              padding: "20px",
              boxShadow: "0px 4px 10px rgba(0, 0, 0, 0.1)",
            }}
          >
            <Input
              variant="outlined"
              placeholder="Paste a link to shorten it"
              fullWidth
              value={originalUrl}
              onChange={(e) => setOriginalUrl(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <LinkIcon />
                  </InputAdornment>
                ),
              }}
              sx={{
                backgroundColor: "#fff",
                borderRadius: "12px",
                "& .MuiOutlinedInput-root": {
                  borderRadius: "12px",
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
                marginLeft: "10px",
                padding: "10px 20px",
                borderRadius: "12px",
                backgroundColor: "#7f5fff",
                boxShadow: "none",
                ":hover": {
                  backgroundColor: "#6f4eff",
                },
              }}
            >
              Shorten
            </Button>
          </Box>
        </Box>
        {shortUrl && (
          <Typography
            variant="h6"
            color="textSecondary"
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
      <Box textAlign="center">
        <Typography fontWeight="600" fontSize="2rem" color="rgb(48,42,150)">
          A short link, infinite possibilites
        </Typography>
        <Typography fontSize="1.5rem">
          With the advanced intelligent link shortening service, you can
          customize your links and share them easily
        </Typography>
      </Box>
      <Box display="flex" justifyContent="space-evenly">
        <Box
          textAlign="center"
          border="solid 1px black"
          borderRadius="20px"
          maxWidth="400px"
        >
          <Typography fontWeight="600" fontSize="1.5rem" color="rgb(48,42,150)">
            Custom Domains
          </Typography>
          <Typography>
            Track audience individually for each brand,website or client by
            using your own domain or subdomain for link shortening.
          </Typography>
        </Box>
        <Box
          textAlign="center"
          border="solid 1px black"
          borderRadius="20px"
          maxWidth="400px"
        >
          <Typography fontWeight="600" fontSize="1.5rem" color="rgb(48,42,150)">
            Track Clicks
          </Typography>
          <Typography>
            Focus your or your client's efforts on the most promising campaigns
            by taking actions based on comprehensive statistics.
          </Typography>
        </Box>
        <Box
          textAlign="center"
          border="solid 1px black"
          borderRadius="20px"
          maxWidth="400px"
        >
          <Typography fontWeight="600" fontSize="1.5rem" color="rgb(48,42,150)">
            Friendly Support
          </Typography>
          <Typography>
            We really care about your success in using short links, so you
            always get answers.
          </Typography>
        </Box>
      </Box>
    </Box>
  );
}

export default App;
