import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RiskGauge } from "@/components/analyst/RiskGauge";

describe("smoke", () => {
  it("renders button", () => {
    render(<Button>Hello</Button>);
    expect(screen.getByText("Hello")).toBeTruthy();
  });

  it("renders badge", () => {
    render(<Badge tone="success">OK</Badge>);
    expect(screen.getByText("OK")).toBeTruthy();
  });

  it("renders risk gauge", () => {
    render(<RiskGauge score={650} band="HIGH" />);
    expect(screen.getByText("Risk score")).toBeTruthy();
  });
});
