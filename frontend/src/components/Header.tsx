import { Box } from "@mui/material";
import "../App.css";

const Header = () => {
  return (
    <Box>
      <h1 className="text-4xl font-bold text-center text-indigo-700 mb-4">
        Create Short Links!
      </h1>
      <p className="text-xl text-center mb-8">
        _ is a custom short link personalization tool that enables you to
        target, engage, and drive more customers. Get started now.
      </p>
    </Box>
  );
};

export default Header;
