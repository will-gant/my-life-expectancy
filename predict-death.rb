require 'csv'
require 'date'
require 'set'

def compose_message(years)
  years.negative? ? "#{years.abs} fewer years" : "#{years.abs} more years"
end

ancestors = CSV.read('direct-ancestors.csv', headers: true)
female_death_stats_csv = CSV.open('female_death_stats.csv', headers: true, header_converters: :symbol)
male_death_stats_csv = CSV.open('male_death_stats.csv', headers: true, header_converters: :symbol)

uk_death_stats = { male: [], female: [] }
male_death_stats_csv.each do |row|
  row_hash = row.to_h
  uk_death_stats[:male] << row_hash
end
female_death_stats_csv.each do |row|
  row_hash = row.to_h
  uk_death_stats[:female] << row_hash
end
female_death_stats_csv.close
male_death_stats_csv.close

individuals_with_age_at_death = { male: [], female: [] }

ancestors.each do |row|
  gender = row['Gender']&.downcase
  next if row['Birth date'].nil? || row['Death date'].nil? || gender.empty?

  begin
    birth_date = Date.parse(row['Birth date'])
  rescue ArgumentError => e
    next
  end

  begin
    death_date = Date.parse(row['Death date'])
  rescue ArgumentError => e
    next
  end

  relevant_death_stats = uk_death_stats[gender.to_sym].detect { |stat| stat[:year].to_i == death_date.year }

  # skip if we have no UK stats for the year of death
  next if relevant_death_stats.nil?

  age_at_death = (death_date - birth_date).to_f / 365.0

  # exclude childhood deaths on the same basis as the UK's ONS does
  next unless age_at_death >= 10

  individuals_with_age_at_death[gender.to_sym] << {
    year_of_death: death_date.year,
    age_at_death: age_at_death,
    modal_death_diff: age_at_death - relevant_death_stats[:modal_age_at_death].to_f,
    median_death_diff: age_at_death - relevant_death_stats[:median_age_at_death].to_f,
    life_expectancy_at_birth_diff: age_at_death - relevant_death_stats[:life_expectancy_at_birth].to_f
  }
end

diff = { male: {}, female: {} }
Set[:male, :female].each do |gender|
  Set[:modal_death_diff, :median_death_diff, :life_expectancy_at_birth_diff].each do |diff_type|
    diff[gender][diff_type] = (individuals_with_age_at_death[gender].reduce(0.0) do |sum, person|
      sum + person[diff_type]
    end / individuals_with_age_at_death[gender].size).round(2)
  end
end

diff.each do |gender, stats|
  ancestor_count = individuals_with_age_at_death[gender].size
  earliest_death = individuals_with_age_at_death[gender].sort_by { |ancestor| ancestor[:year_of_death] }.first[:year_of_death]
  latest_death = individuals_with_age_at_death[gender].sort_by { |ancestor| ancestor[:year_of_death] }.reverse.first[:year_of_death]

  puts "#{gender} ancestors in the provided dataset lived #{compose_message(stats[:modal_death_diff])} than the UK's modal age of death in the year they died"
  puts "#{gender} ancestors in the provided dataset lived #{compose_message(stats[:median_death_diff])} than the UK's median age of death in the year they died"
  puts "Calculated from #{ancestor_count} #{gender.to_s} ancestors who died between #{earliest_death} and #{latest_death}"
end