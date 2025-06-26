async function act(kind) {
  const res = await fetch("/api/" + kind, {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: src.value,
  });
  const data = await res.json();
  out.textContent = data.out || data.listing || "";
  log.textContent = data.log || "";
}
