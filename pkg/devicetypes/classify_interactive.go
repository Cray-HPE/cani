package devicetypes

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ClassifyOptions controls the interactive classification prompt.
type ClassifyOptions struct {
	NoColor bool
	Writer  io.Writer // defaults to os.Stdout
	Reader  io.Reader // defaults to os.Stdin
}

func (o ClassifyOptions) writer() io.Writer {
	if o.Writer != nil {
		return o.Writer
	}
	return os.Stdout
}

func (o ClassifyOptions) reader() io.Reader {
	if o.Reader != nil {
		return o.Reader
	}
	return os.Stdin
}

// PromptForDeviceType displays an interactive menu that lets the user pick a
// device-type slug for an unclassified device. Returns the selected slug or
// an empty string if the user chose to skip.
func PromptForDeviceType(device UnclassifiedDevice, opts ClassifyOptions) (string, error) {
	w := opts.writer()
	r := bufio.NewReader(opts.reader())

	cyan, yellow, green, gray, bold := colorFuncs(opts.NoColor)

	// Display device summary
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", bold("─── Unclassified Device ───"))
	fmt.Fprintf(w, "  Name:          %s\n", cyan(device.Name))
	if device.DeviceType != "" {
		fmt.Fprintf(w, "  Type:          %s\n", device.DeviceType)
	}
	if device.Status != "" {
		fmt.Fprintf(w, "  Status:        %s\n", device.Status)
	}
	if device.Role != "" {
		fmt.Fprintf(w, "  Role:          %s\n", device.Role)
	}
	if device.DeviceType != "" {
		fmt.Fprintf(w, "  Type:          %s\n", device.DeviceType)
	}
	if device.Model != "" {
		fmt.Fprintf(w, "  Model:         %s\n", device.Model)
	}
	if device.Manufacturer != "" {
		fmt.Fprintf(w, "  Manufacturer:  %s\n", device.Manufacturer)
	}
	if device.ChildrenCount > 0 {
		fmt.Fprintf(w, "  Children:      %d\n", device.ChildrenCount)
	}
	printProviderMetadata(w, device.ProviderMetadata, gray)

	// Show suggestions
	suggestions := SuggestTypes(device, 8)
	if len(suggestions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", yellow("Suggestions:"))
		for i, s := range suggestions {
			label := ScoreTierLabel(s.Score)
			fmt.Fprintf(w, "  %s  %s  %s\n",
				bold(fmt.Sprintf("[%d]", i+1)),
				s.Slug,
				gray(fmt.Sprintf("(%d%% %s)", s.Score, label)),
			)
		}
	} else {
		fmt.Fprintf(w, "\n%s\n", gray("  No suggestions found"))
	}

	// Show options
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s select suggestion, %s to search, %s to skip\n",
		bold("[1-N]"), bold("[s]"), bold("[k]"))
	fmt.Fprintf(w, "  > ")

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		input := strings.TrimSpace(line)

		switch strings.ToLower(input) {
		case "k", "skip":
			return "", nil
		case "s", "search":
			slug, err := promptSearch(w, r, opts.NoColor)
			if err != nil {
				return "", err
			}
			if slug != "" {
				fmt.Fprintf(w, "  %s %s\n", green("✓"), slug)
			}
			return slug, nil
		default:
			n, err := strconv.Atoi(input)
			if err == nil && n >= 1 && n <= len(suggestions) {
				slug := suggestions[n-1].Slug
				fmt.Fprintf(w, "  %s %s\n", green("✓"), slug)
				return slug, nil
			}
			// Try as a direct slug
			if _, ok := GetBySlug(input); ok {
				fmt.Fprintf(w, "  %s %s\n", green("✓"), input)
				return input, nil
			}
			fmt.Fprintf(w, "  Invalid input. Enter a number, 's' to search, or 'k' to skip: ")
		}
	}
}

// promptSearch presents a free-text search prompt and returns the selected slug.
func promptSearch(w io.Writer, r *bufio.Reader, noColor bool) (string, error) {
	_, yellow, _, gray, bold := colorFuncs(noColor)

	fmt.Fprintf(w, "  %s ", yellow("Search query:"))
	line, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading search input: %w", err)
	}
	query := strings.TrimSpace(line)
	if query == "" {
		return "", nil
	}

	results := searchSlugs(query, 10)
	if len(results) == 0 {
		fmt.Fprintf(w, "  %s\n", gray("No matches found"))
		return "", nil
	}

	fmt.Fprintf(w, "  %s\n", yellow("Search results:"))
	for i, slug := range results {
		fmt.Fprintf(w, "    %s  %s\n", bold(fmt.Sprintf("[%d]", i+1)), slug)
	}
	fmt.Fprintf(w, "  Select or 'k' to skip: ")

	line, err = r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading selection: %w", err)
	}
	input := strings.TrimSpace(line)
	if strings.ToLower(input) == "k" {
		return "", nil
	}
	n, err := strconv.Atoi(input)
	if err == nil && n >= 1 && n <= len(results) {
		return results[n-1], nil
	}
	return "", nil
}

// searchSlugs searches all registered slugs for substring matches, returning
// up to maxResults. Case-insensitive.
func searchSlugs(query string, maxResults int) []string {
	lower := strings.ToLower(query)
	var matches []string
	for _, slug := range GetAllSlugs() {
		if strings.Contains(strings.ToLower(slug), lower) {
			matches = append(matches, slug)
			if len(matches) >= maxResults {
				break
			}
		}
	}
	return matches
}

// colorFuncs returns coloring closures, respecting noColor.
func colorFuncs(noColor bool) (cyan, yellow, green, gray, bold func(string) string) {
	id := func(s string) string { return s }
	if noColor {
		return id, id, id, id, id
	}
	cyan = func(s string) string { return "\033[36m" + s + "\033[0m" }
	yellow = func(s string) string { return "\033[33m" + s + "\033[0m" }
	green = func(s string) string { return "\033[32m" + s + "\033[0m" }
	gray = func(s string) string { return "\033[90m" + s + "\033[0m" }
	bold = func(s string) string { return "\033[1m" + s + "\033[0m" }
	return
}

// printProviderMetadata writes a compact representation of provider metadata.
func printProviderMetadata(w io.Writer, pm map[string]any, gray func(string) string) {
	if len(pm) == 0 {
		return
	}
	for provider, raw := range pm {
		meta, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		// Display selected useful keys in a compact line.
		var parts []string
		for _, key := range []string{"xname", "class", "role", "subRole", "nid", "state", "aliases"} {
			v, exists := meta[key]
			if !exists {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s=%v", key, v))
		}
		if len(parts) > 0 {
			fmt.Fprintf(w, "  %s:  %s\n",
				provider,
				gray(strings.Join(parts, ", ")),
			)
		}
	}
}
