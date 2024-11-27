package redis

type Config struct{
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Redis struct{

}

func New(cfg *Config) (*Redis, error){
	
}