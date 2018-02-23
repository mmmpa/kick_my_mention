require 'yaml'

class Combine
  CONFIG_PATH = '.circleci/config.yml'
  CONFIG_BASE_PATH = '.circleci/_config.yml'
  CONFIG_CHILDREN_PATH = '.circleci/**/*.yml'

  def execute!
    check!
    prepare!

    File.write(
      CONFIG_PATH,
      base_configuration.merge('jobs' => combined_configuration).to_yaml
    )

    commit!
  rescue ConfigurationHasChange => e
    puts "\e[31m#{e.message}\e[0m"
    exit(1)
  end

  private

  def check!
    return unless File.exist?(CONFIG_PATH)
    raise ConfigurationHasChange if config_yml_has_change?
  end

  def config_yml_has_change?
    return false unless `git status`.match?(/modified:.+#{CONFIG_PATH}/)

    `git diff HEAD^ -- #{CONFIG_PATH}` != ''
  end

  def prepare!
    File.delete(CONFIG_PATH)
  rescue => e
    puts "#{e} (But ignore this error.)"
  end

  def base_configuration
    YAML.load_file(CONFIG_BASE_PATH)
  end

  def configurations
    Dir.glob(CONFIG_CHILDREN_PATH)
  end

  def pick_name(f)
    File.basename(f).split('.').shift
  end

  def combined_configuration
    configurations.inject({}) do |a, f|
      name = pick_name(f)

      if name == '_config'
        a
      else
        a.merge(name => YAML.load_file(f)['jobs']['build'])
      end
    end
  end

  def commit!
    `git add #{CONFIG_PATH}`
  end

  class ConfigurationHasChange < RuntimeError
    def message
      <<-EOS
"#{CONFIG_PATH}" has change.
DO NOT update "#{CONFIG_PATH}" manually.
Any configuration must be split out to "\#{BUILD_NAME}.yml".
      EOS
    end
  end
end

Combine.new.execute! if $0 == __FILE__
