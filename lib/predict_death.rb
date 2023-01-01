require 'csv'
require 'date'
require 'optparse'

SECONDS_IN_NON_LEAP_YEAR = 31536000
YEAR_REGEX = /\b(18|19|20)\d{2}\b/

def compose_message(years)
  years.negative? ? "#{years.abs} fewer years" : "#{years.abs} more years"
end

def compile_uk_death_stats(male_file, female_file)
  female_death_stats_csv = CSV.open(female_file, headers: true, header_converters: :symbol)
  male_death_stats_csv = CSV.open(male_file, headers: true, header_converters: :symbol)

  uk_death_stats = { male: [], female: [] }
  male_death_stats_csv.each do |row|
    row_hash = row.to_h
    uk_death_stats[:male] << row_hash
  end
  female_death_stats_csv.each do |row|
    row_hash = row.to_h
    uk_death_stats[:female] << row_hash
  end
  uk_death_stats
end

def sufficient_data?(ancestor)
  !ancestor['Birth date'].nil? && !ancestor['Death date'].nil? && !ancestor['Gender'].empty?
end

def find_relevant_death_stats(death_date, gender, uk_death_stats)
  uk_death_stats[gender.to_sym].detect { |stat| stat[:year].to_i == death_date.year }
end

def parse_date(date)
  begin
    Date.parse(date)
  rescue ArgumentError => e
    year_match = date.match(YEAR_REGEX)
    return if year_match.nil?

    Date.strptime(year_match[0], "%Y")
  end
end

def extract_ancestor_death_details(ancestors, uk_death_stats)
  ancestors_with_age_at_death = { male: [], female: [] }

  ancestors.each do |ancestor|
    gender = ancestor['Gender']&.downcase
    next unless sufficient_data?(ancestor)

    birth_date = parse_date(ancestor['Birth date'])
    next if birth_date.nil?

    death_date = parse_date(ancestor['Death date'])
    next if death_date.nil?

    relevant_death_stats = find_relevant_death_stats(death_date, gender, uk_death_stats)
  
    # skip if we have no UK stats for the year of death
    next if relevant_death_stats.nil?
  
    age_at_death_seconds = death_date.to_time.to_i - birth_date.to_time.to_i
  
    # exclude childhood deaths on the same basis as the UK's ONS does
    next unless age_at_death_seconds >= 10 * SECONDS_IN_NON_LEAP_YEAR

    ancestors_with_age_at_death[gender.to_sym] << {
      year_of_death: death_date.year,
      age_at_death: (age_at_death_seconds.to_i / SECONDS_IN_NON_LEAP_YEAR.to_i),
      modal_diff: age_at_death_seconds - relevant_death_stats[:modal_age_at_death].to_f * SECONDS_IN_NON_LEAP_YEAR,
      median_diff: age_at_death_seconds - relevant_death_stats[:median_age_at_death].to_f * SECONDS_IN_NON_LEAP_YEAR
    }
  end

  ancestors_with_age_at_death
end

def average(ancestors, diff_type)
  (ancestors.reduce(0.0) do |sum, ancestor|
    sum + ancestor[diff_type]
  end / ancestors.size).round(2)
end

def calculate_death_diff_stats(ancestor_death_details)
  diff = { male: {}, female: {} }
  [:male, :female].each do |gender|
    [:modal_diff, :median_diff].each do |diff_type|
      diff[gender][diff_type] = (average(ancestor_death_details[gender], diff_type) / SECONDS_IN_NON_LEAP_YEAR).round(2)
    end
  end
  diff
end

options = {
  male_death_stats: File.join(__dir__, '../male_death_stats.csv'),
  female_death_stats: File.join(__dir__, '../female_death_stats.csv'),
  ancestors: File.join(__dir__, '../direct-ancestors.csv')
}

OptionParser.new do |opts|
  opts.on('--male-death-stats FILE') { |file| options[:male_death_stats] = file }
  opts.on('--female-death-stats FILE') { |file| options[:female_death_stats] = file }
  opts.on('--ancestors FILE') { |file| options[:ancestors] = file }
end.parse!

ancestors = CSV.read(options[:ancestors], headers: true)

ancestor_death_details = extract_ancestor_death_details(ancestors, compile_uk_death_stats(options[:male_death_stats], options[:female_death_stats]))

diff = calculate_death_diff_stats(ancestor_death_details)

diff.each do |gender, stats|
  ancestor_count = ancestor_death_details[gender].size
  earliest_death = ancestor_death_details[gender].sort_by { |ancestor| ancestor[:year_of_death] }.first[:year_of_death]
  latest_death = ancestor_death_details[gender].sort_by { |ancestor| ancestor[:year_of_death] }.reverse.first[:year_of_death]
  count_outlived_median = ancestor_death_details[gender].count { |ancestor| ancestor[:median_diff].positive? }
  count_outlived_modal = ancestor_death_details[gender].count { |ancestor| ancestor[:modal_diff].positive? }

  puts
  puts "#{gender} ancestors in the provided dataset lived #{compose_message(stats[:modal_diff])} than the UK's #{gender} modal age of death in the year they died (#{count_outlived_modal}/#{ancestor_count} outlived the mode)"
  puts "#{gender} ancestors in the provided dataset lived #{compose_message(stats[:median_diff])} than the UK's #{gender} median age of death in the year they died (#{count_outlived_median}/#{ancestor_count} outlived the median)"
  puts "Calculated from #{ancestor_count} #{gender.to_s} ancestors who died between #{earliest_death} and #{latest_death}"
end