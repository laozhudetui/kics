package console

import (
	_ "embed" // Embed kics CLI img
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	consoleHelpers "github.com/Checkmarx/kics/internal/console/helpers"
	"github.com/Checkmarx/kics/internal/storage"
	"github.com/Checkmarx/kics/internal/tracker"
	"github.com/Checkmarx/kics/pkg/engine"
	"github.com/Checkmarx/kics/pkg/engine/provider"
	"github.com/Checkmarx/kics/pkg/engine/source"
	"github.com/Checkmarx/kics/pkg/kics"
	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser"
	dockerParser "github.com/Checkmarx/kics/pkg/parser/docker"
	jsonParser "github.com/Checkmarx/kics/pkg/parser/json"
	terraformParser "github.com/Checkmarx/kics/pkg/parser/terraform"
	yamlParser "github.com/Checkmarx/kics/pkg/parser/yaml"
	"github.com/Checkmarx/kics/pkg/resolver"
	"github.com/Checkmarx/kics/pkg/resolver/helm"
	"github.com/getsentry/sentry-go"
	"github.com/gookit/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	path              string
	queryPath         string
	outputPath        string
	payloadPath       string
	excludeCategories []string
	excludePath       []string
	excludeIDs        []string
	excludeResults    []string
	reportFormats     []string
	cfgFile           string
	verbose           bool
	logFile           bool
	noProgress        bool
	types             []string
	noColor           bool
	min               bool
	outputLines       int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Executes a scan analysis",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return scan()
	},
}

var listPlatformsCmd = &cobra.Command{
	Use:   "list-platforms",
	Short: "List supported platforms",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, v := range source.ListSupportedPlatforms() {
			fmt.Println(v)
		}
		return nil
	},
}

func initializeConfig(cmd *cobra.Command) error {
	if cfgFile == "" {
		configpath := path
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			configpath = filepath.Dir(path)
		}
		_, err = os.Stat(filepath.ToSlash(filepath.Join(configpath, "kics.config")))
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		cfgFile = filepath.ToSlash(filepath.Join(path, "kics.config"))
	}
	v := viper.New()
	base := filepath.Base(cfgFile)
	v.SetConfigName(base)
	v.AddConfigPath(filepath.Dir(cfgFile))
	ext, err := consoleHelpers.FileAnalyzer(cfgFile)
	if err != nil {
		return err
	}
	v.SetConfigType(ext)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.SetEnvPrefix("KICS_")
	v.AutomaticEnv()
	bindFlags(cmd, v)
	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	settingsMap := v.AllSettings()
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		settingsMap[f.Name] = true
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if err := v.BindEnv(f.Name, fmt.Sprintf("%s_%s", "KICS", envVarSuffix)); err != nil {
				log.Err(err).Msg("Failed to bind Viper flags")
			}
		}
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			setBoundFlags(f.Name, val, cmd)
		}
	})
	for key, val := range settingsMap {
		if val == true {
			continue
		} else {
			fmt.Printf("Unknown configuration key: '%s'\nShowing help for '%s' command:\n\n", key, cmd.Name())
			err := cmd.Help()
			if err != nil {
				log.Err(err).Msg("Unable to show help message")
			}
			os.Exit(1)
		}
	}
}

func setBoundFlags(flagName string, val interface{}, cmd *cobra.Command) {
	switch t := val.(type) {
	case []interface{}:
		var paramSlice []string
		for _, param := range t {
			paramSlice = append(paramSlice, param.(string))
		}
		valStr := strings.Join(paramSlice, ",")
		if err := cmd.Flags().Set(flagName, fmt.Sprintf("%v", valStr)); err != nil {
			log.Err(err).Msg("Failed to get Viper flags")
		}
	default:
		if err := cmd.Flags().Set(flagName, fmt.Sprintf("%v", val)); err != nil {
			log.Err(err).Msg("Failed to get Viper flags")
		}
	}
}

func initScanCmd() {
	scanCmd.Flags().StringVarP(&path, "path", "p", "", "path or directory path to scan")
	scanCmd.Flags().StringVarP(&cfgFile, "config", "", "", "path to configuration file")
	scanCmd.Flags().StringVarP(
		&queryPath,
		"queries-path",
		"q",
		"./assets/queries",
		"path to directory with queries",
	)
	scanCmd.Flags().StringVarP(&outputPath, "output-path", "o", "", "directory path to store reports")
	scanCmd.Flags().StringSliceVarP(
		&reportFormats,
		"report-formats",
		"",
		[]string{},
		"formats in which the results will be exported (json, sarif, html)",
	)
	scanCmd.Flags().IntVarP(&outputLines, "preview-lines", "", 3, "number of lines to be display in CLI results (default: 3)")
	scanCmd.Flags().StringVarP(&payloadPath, "payload-path", "d", "", "path to store internal representation JSON file")
	scanCmd.Flags().StringSliceVarP(
		&excludePath,
		"exclude-paths",
		"e",
		[]string{},
		"exclude paths from scan\nsupports glob and can be provided multiple times or as a quoted comma separated string"+
			"\nexample: './shouldNotScan/*,somefile.txt'",
	)
	scanCmd.Flags().BoolVarP(&noColor, "no-color", "", false, "disable CLI color output")
	scanCmd.Flags().BoolVarP(&min, "minimal-ui", "", false, "simplified version of CLI output")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "increase verbosity")
	scanCmd.Flags().BoolVarP(&logFile, "log-file", "l", false, "writes log messages to info.log")
	scanCmd.Flags().StringSliceVarP(&types, "type", "t", []string{""}, "case insensitive list of platform types to scan\n"+
		fmt.Sprintf("(%s)", strings.Join(source.ListSupportedPlatforms(), ", ")))
	scanCmd.Flags().BoolVarP(&noProgress, "no-progress", "", false, "hides the progress bar")
	scanCmd.Flags().StringSliceVarP(
		&excludeIDs,
		"exclude-queries",
		"",
		[]string{},
		"exclude queries by providing the query ID\n"+
			"can be provided multiple times or as a comma separated string\n"+
			"example: 'e69890e6-fce5-461d-98ad-cb98318dfc96,4728cd65-a20c-49da-8b31-9c08b423e4db'",
	)
	scanCmd.Flags().StringSliceVarP(
		&excludeResults,
		"exclude-results",
		"x",
		[]string{},
		"exclude results by providing the similarity ID of a result\n"+
			"can be provided multiple times or as a comma separated string\n"+
			"example: 'fec62a97d569662093dbb9739360942f...,31263s5696620s93dbb973d9360942fc2a...'",
	)
	scanCmd.Flags().StringSliceVarP(
		&excludeCategories,
		"exclude-categories",
		"",
		[]string{},
		"exclude categories by providing its name\n"+
			"can be provided multiple times or as a comma separated string\n"+
			"example: 'Access control,Best practices'",
	)

	if err := scanCmd.MarkFlagRequired("path"); err != nil {
		sentry.CaptureException(err)
		log.Err(err).Msg("failed to add command required flags")
	}

	scanCmd.AddCommand(listPlatformsCmd)
}

func setupLogs() error {
	consoleLogger := zerolog.ConsoleWriter{Out: io.Discard}
	fileLogger := zerolog.ConsoleWriter{Out: io.Discard}

	if verbose {
		consoleLogger = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	if logFile {
		file, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		fileLogger = consoleHelpers.CustomConsoleWriter(&zerolog.ConsoleWriter{Out: file, NoColor: true})
	}

	mw := io.MultiWriter(consoleLogger, fileLogger)
	log.Logger = log.Output(mw)
	return nil
}

func getFileSystemSourceProvider() (*provider.FileSystemSourceProvider, error) {
	var excludePaths []string
	if payloadPath != "" {
		excludePaths = append(excludePaths, payloadPath)
	}

	if len(excludePath) > 0 {
		excludePaths = append(excludePaths, excludePath...)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	filesSource, err := provider.NewFileSystemSourceProvider(absPath, excludePaths)
	if err != nil {
		return nil, err
	}
	return filesSource, nil
}

func getExcludeResultsMap(excludeResults []string) map[string]bool {
	excludeResultsMap := make(map[string]bool)
	for _, er := range excludeResults {
		excludeResultsMap[er] = true
	}
	return excludeResultsMap
}

//go:embed img/kics-console
var s string

func scan() error { //nolint
	if noColor {
		color.Disable()
	}

	printer := consoleHelpers.NewPrinter(min)
	printer.Success.Printf("\n%s\n\n", s)
	fmt.Printf("Scanning with %s\n\n", getVersion())

	if errlog := setupLogs(); errlog != nil {
		return errlog
	}
	scanStartTime := time.Now()

	querySource := source.NewFilesystemSource(queryPath, types)

	t, err := tracker.NewTracker(outputLines)
	if err != nil {
		return err
	}

	excludeResultsMap := getExcludeResultsMap(excludeResults)

	excludeQueries := source.ExcludeQueries{
		ByIDs:        excludeIDs,
		ByCategories: excludeCategories,
	}

	inspector, err := engine.NewInspector(ctx, querySource, engine.DefaultVulnerabilityBuilder, t, excludeQueries, excludeResultsMap)
	if err != nil {
		return err
	}

	filesSource, err := getFileSystemSourceProvider()
	if err != nil {
		return err
	}

	combinedParser, err := parser.NewBuilder().
		Add(&jsonParser.Parser{}).
		Add(&yamlParser.Parser{}).
		Add(terraformParser.NewDefault()).
		Add(&dockerParser.Parser{}).
		Build(querySource.Types)
	if err != nil {
		return err
	}

	// combinedResolver to be used to resolve files and templates
	combinedResolver, err := resolver.NewBuilder().
		Add(&helm.Resolver{}).
		Build()
	if err != nil {
		return err
	}

	store := storage.NewMemoryStorage()

	service := &kics.Service{
		SourceProvider: filesSource,
		Storage:        store,
		Parser:         combinedParser,
		Inspector:      inspector,
		Tracker:        t,
		Resolver:       combinedResolver,
	}

	if scanErr := service.StartScan(ctx, scanID, noProgress); scanErr != nil {
		return scanErr
	}

	result, err := store.GetVulnerabilities(ctx, scanID)
	if err != nil {
		return err
	}

	files, err := store.GetFiles(ctx, scanID)
	if err != nil {
		return err
	}

	elapsed := time.Since(scanStartTime)

	counters := model.Counters{
		ScannedFiles:           t.FoundFiles,
		ParsedFiles:            t.ParsedFiles,
		TotalQueries:           t.LoadedQueries,
		FailedToExecuteQueries: t.LoadedQueries - t.ExecutedQueries,
		FailedSimilarityID:     t.FailedSimilarityID,
	}

	summary := model.CreateSummary(counters, result, scanID)

	if err := resolveOutputs(&summary, files.Combine(), inspector.GetFailedQueries(), printer); err != nil {
		return err
	}

	elapsedStrFormat := "Scan duration: %v\n"
	fmt.Printf(elapsedStrFormat, elapsed)
	log.Info().Msgf(elapsedStrFormat, elapsed)

	if summary.FailedToExecuteQueries > 0 {
		os.Exit(1)
	}

	return nil
}

func resolveOutputs(
	summary *model.Summary,
	documents model.Documents,
	failedQueries map[string]error,
	printer *consoleHelpers.Printer,
) error {
	if err := printOutput(payloadPath, "payload", documents, []string{"json"}); err != nil {
		return err
	}

	if err := printOutput(outputPath, "results", summary, reportFormats); err != nil {
		return err
	}

	return consoleHelpers.PrintResult(summary, failedQueries, printer)
}

func printOutput(outputPath, filename string, body interface{}, formats []string) error {
	if outputPath == "" {
		return nil
	}
	if strings.Contains(outputPath, ".") {
		if len(formats) == 0 && filepath.Ext(outputPath) != "" {
			formats = []string{filepath.Ext(outputPath)[1:]}
		}
		if len(formats) == 1 && strings.HasSuffix(outputPath, formats[0]) {
			filename = filepath.Base(outputPath)
			outputPath = filepath.Dir(outputPath)
		}
	}

	ok := consoleHelpers.ValidateReportFormats(formats)
	if ok == nil {
		ok = consoleHelpers.GenerateReport(outputPath, filename, body, formats)
	}
	return ok
}
