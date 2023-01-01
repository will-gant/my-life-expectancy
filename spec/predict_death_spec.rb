require 'rspec'
require 'predict_death'

describe '#compose_message' do
  it 'returns "5 more years" when given 5' do
    expect(compose_message(5)).to eq("5 more years")
  end

  it 'returns "3 fewer years" when given -3' do
    expect(compose_message(-3)).to eq("3 fewer years")
  end
end

describe '#sufficient_data?' do
  let(:ancestor) { { 'Birth date' => '1900-01-01', 'Death date' => '1950-01-01', 'Gender' => 'Female' } }

  it 'returns true when given valid ancestor data' do
    expect(sufficient_data?(ancestor)).to be true
  end

  it 'returns false when given nil for Birth date' do
    ancestor['Birth date'] = nil
    expect(sufficient_data?(ancestor)).to be false
  end

  it 'returns false when given nil for Death date' do
    ancestor['Death date'] = nil
    expect(sufficient_data?(ancestor)).to be false
  end

  it 'returns false when given an empty string for Gender' do
    ancestor['Gender'] = ''
    expect(sufficient_data?(ancestor)).to be false
  end
end

describe 'find_relevant_death_stats' do
  let(:uk_death_stats) do
    {
      male: [
        { year: '1900', modal_age_at_death: '70' },
        { year: '1901', modal_age_at_death: '75' }
      ],
      female: [
        { year: '1900', modal_age_at_death: '75' },
        { year: '1901', modal_age_at_death: '80' }
      ]
    }
  end

  context 'when the death date is found in the UK death stats' do
    it 'returns the death stats for the given year and gender' do
      death_date = Date.new(1900, 1, 1)
      expect(find_relevant_death_stats(death_date, :male, uk_death_stats)).to eq({ year: '1900', modal_age_at_death: '70' })
      expect(find_relevant_death_stats(death_date, :female, uk_death_stats)).to eq({ year: '1900', modal_age_at_death: '75' })
    end
  end

  context 'when the death date is not found in the UK death stats' do
    it 'returns nil' do
      death_date = Date.new(1902, 1, 1)
      expect(find_relevant_death_stats(death_date, :male, uk_death_stats)).to be_nil
      expect(find_relevant_death_stats(death_date, :female, uk_death_stats)).to be_nil
    end
  end
end

describe '#extract_ancestor_death_details' do
  let(:uk_death_stats) do
    {
      male: [
        { year: '1970', modal_age_at_death: '72.0', median_age_at_death: '73.0', life_expectancy_at_birth: '70.0' },
        { year: '1971', modal_age_at_death: '73.0', median_age_at_death: '74.0', life_expectancy_at_birth: '71.0' }
      ],
      female: [
        { year: '1970', modal_age_at_death: '79.0', median_age_at_death: '80.0', life_expectancy_at_birth: '76.0' },
        { year: '1971', modal_age_at_death: '80.0', median_age_at_death: '81.0', life_expectancy_at_birth: '77.0' }
      ]
    }
  end

  it 'extracts ancestor death details' do
    ancestors = [
      { 'Birth date' => '1900-01-01', 'Death date' => '1970-06-01', 'Gender' => 'Male' },
      { 'Birth date' => '1901-01-01', 'Death date' => '1971', 'Gender' => 'Male' },
      { 'Birth date' => '1871-01-01', 'Death date' => '1971-01-01', 'Gender' => 'Female' },
      { 'Birth date' => '1880-01-01', 'Death date' => '1 January 1970', 'Gender' => 'Female' },
      { 'Birth date' => '1870-01-01', 'Death date' => 'about 1970', 'Gender' => 'Female' }
    ]
    expected_output = {
      male: [
        { year_of_death: 1970, age_at_death: 70, modal_diff: -48556800.0, median_diff: -80092800.0 },
        { year_of_death: 1971, age_at_death: 70, modal_diff: -93139200.0, median_diff: -124675200.0 }
      ],
      female: [
        { year_of_death: 1971, age_at_death: 100, modal_diff: 632793600.0, median_diff: 601257600.0 },
        { year_of_death: 1970, age_at_death: 90, modal_diff: 348796800.0, median_diff: 317260800.0 },
        { year_of_death: 1970, age_at_death: 100, modal_diff: 664329600.0, median_diff: 632793600.0 }
      ]
    }

    expect(extract_ancestor_death_details(ancestors, uk_death_stats)).to eq(expected_output)
  end

  it 'filters out ancestors with missing data' do
    ancestors = [
      { 'Birth date' => '1920-01-01', 'Death date' => '', 'Gender' => 'Male' },
      { 'Birth date' => '', 'Death date' => '1995-01-01', 'Gender' => 'Female' }
    ]
    expected_output = { male: [], female: [] }

    expect(extract_ancestor_death_details(ancestors, uk_death_stats)).to eq(expected_output)
  end

  it 'filters out ancestors with invalid birth or death dates' do
    uk_death_stats = {
      male: [
        { year: '1970', modal_age_at_death: '70', median_age_at_death: '71', life_expectancy_at_birth: '45' },
      ],
      female: [
        { year: '1971', modal_age_at_death: '76', median_age_at_death: '77', life_expectancy_at_birth: '51' }
      ]
    }

    ancestors = [
      { 'Birth date' => '1900-01-01', 'Death date' => '1970-01-01', 'Gender' => 'Male' },
      { 'Birth date' => '1901-01-01', 'Death date' => '1971-01-01', 'Gender' => 'Female' },
      { 'Birth date' => 'invalid', 'Death date' => '1971', 'Gender' => 'Female' },
      { 'Birth date' => '1900-01-01', 'Death date' => 'invalid', 'Gender' => 'Female' },
      { 'Birth date' => '1900-01-01', 'Death date' => '1971-01-01', 'Gender' => '' },
      { 'Birth date' => '1500-01-01', 'Death date' => '1571-01-01', 'Gender' => 'Male' }
    ]

    result = extract_ancestor_death_details(ancestors, uk_death_stats)
    expect(result[:male].size).to eq(1)
    expect(result[:male][0][:year_of_death]).to eq(1970)
    expect(result[:female].size).to eq(1)
    expect(result[:female][0][:year_of_death]).to eq(1971)
  end
end

describe '#calculate_death_diff_stats' do
  male_ancestors = [
    { modal_diff: 5 * SECONDS_IN_NON_LEAP_YEAR, median_diff: -3 * SECONDS_IN_NON_LEAP_YEAR },
    { modal_diff: 6 * SECONDS_IN_NON_LEAP_YEAR, median_diff: -4 * SECONDS_IN_NON_LEAP_YEAR },
    { modal_diff: 7 * SECONDS_IN_NON_LEAP_YEAR, median_diff: -5 * SECONDS_IN_NON_LEAP_YEAR }
  ]

  female_ancestors = [
    { modal_diff: 4 * SECONDS_IN_NON_LEAP_YEAR, median_diff: 1.5 * SECONDS_IN_NON_LEAP_YEAR },
    { modal_diff: 5 * SECONDS_IN_NON_LEAP_YEAR, median_diff: 1.75 * SECONDS_IN_NON_LEAP_YEAR },
    { modal_diff: 6 * SECONDS_IN_NON_LEAP_YEAR, median_diff: 2 * SECONDS_IN_NON_LEAP_YEAR }
  ]

  ancestor_death_details = {
    male: male_ancestors,
    female: female_ancestors
  }

  it 'calculates the average modal and median diffs for ancestors' do
    result = calculate_death_diff_stats(ancestor_death_details)
    expect(result[:male][:modal_diff]).to eq(6.0)
    expect(result[:male][:median_diff]).to eq(-4.0)
    expect(result[:female][:modal_diff]).to eq(5.0)
    expect(result[:female][:median_diff]).to eq(1.75)
  end
end

describe '#compile_uk_death_stats' do
  it 'parses the male and female death stats CSV files' do
    male_stats_fixture = File.join(__dir__, 'fixtures/male_death_stats.csv')
    female_stats_fixture = File.join(__dir__, 'fixtures/female_death_stats.csv')
    expected_result = {
      male: [{ year: '2000', modal_age_at_death: '80', median_age_at_death: '82', life_expectancy_at_birth: '75' },
             { year: '2001', modal_age_at_death: '81', median_age_at_death: '83', life_expectancy_at_birth: '76' }],
      female: [{ year: '2000', modal_age_at_death: '85', median_age_at_death: '87', life_expectancy_at_birth: '80' },
               { year: '2001', modal_age_at_death: '86', median_age_at_death: '88', life_expectancy_at_birth: '81' }]
    }
    expect(compile_uk_death_stats(male_stats_fixture, female_stats_fixture)).to eq(expected_result)
  end
end
